package deploy

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/redact"
)

// PrintBanner writes discovery and identity context to stderr.
func PrintBanner(w io.Writer, doc *Document, disc Discovery, kcfg kube.Config, id IdentityInfo) {
	server := id.Server
	if server == "" {
		server, _ = kube.ServerURL(kcfg)
		server = normalizeServer(server)
	}
	fmt.Fprintf(w, "Cluster:    %s  %s\n", doc.NamePrefix, server)
	fmt.Fprintf(w, "Kubeconfig: %s (%s)\n", disc.KubeconfigPath, disc.KubeconfigFrom)
	if disc.OutputsPath != "" {
		fmt.Fprintf(w, "Outputs:    %s (%s)\n", disc.OutputsPath, disc.OutputsFrom)
	}
	fmt.Fprintln(w, id.Line())
}

type planJSONRow struct {
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	Action      string `json:"action"`
	FromVersion string `json:"from_version,omitempty"`
	ToVersion   string `json:"to_version,omitempty"`
	Note        string `json:"note,omitempty"`
}

type planSummary struct {
	Install   int `json:"install"`
	Upgrade   int `json:"upgrade"`
	Adopt     int `json:"adopt"`
	Unchanged int `json:"unchanged"`
	Remove    int `json:"remove"`
	Refused   int `json:"refused"`
	Stuck     int `json:"stuck"`
}

// PlanSchemaVersion is the `plan -o json` document version.
const PlanSchemaVersion = 1

type planJSONOut struct {
	SchemaVersion int           `json:"schema_version"`
	HasChange     bool          `json:"has_change"`
	HasRefused    bool          `json:"has_refused"`
	Summary       planSummary   `json:"summary"`
	Rows          []planJSONRow `json:"rows"`
	Removals      []planJSONRow `json:"removals"`
}

func rowToJSON(r plan.Row) planJSONRow {
	return planJSONRow{
		Namespace: r.Namespace, Name: r.Name, Action: string(r.Action),
		FromVersion: r.FromVersion, ToVersion: r.ToVersion, Note: r.Note,
	}
}

func summarizePlan(res plan.Result) planSummary {
	var s planSummary
	removalKeys := map[string]bool{}
	for _, r := range res.Removals {
		removalKeys[r.Namespace+"/"+r.Name] = true
	}
	count := func(row plan.Row) {
		switch row.Action {
		case plan.ActionInstall:
			s.Install++
		case plan.ActionUpgrade:
			s.Upgrade++
		case plan.ActionAdopt:
			s.Adopt++
		case plan.ActionUnchanged:
			s.Unchanged++
		case plan.ActionRemove:
			s.Remove++
		case plan.ActionRefused:
			s.Refused++
		case plan.ActionStuck:
			s.Stuck++
		}
	}
	for _, row := range res.Rows {
		if removalKeys[row.Namespace+"/"+row.Name] {
			continue
		}
		count(row)
	}
	for _, row := range res.Removals {
		count(row)
	}
	return s
}

// PrintPlan renders a plan to stdout as grouped text or JSON.
func PrintPlan(w io.Writer, rd *redact.Redactor, res plan.Result, asJSON bool) error {
	if asJSON {
		out := planJSONOut{
			SchemaVersion: PlanSchemaVersion,
			HasChange:     res.HasChange, HasRefused: planHasRefused(res),
			Summary: summarizePlan(res),
		}
		for _, row := range res.Rows {
			if row.Action == plan.ActionRemove {
				continue
			}
			out.Rows = append(out.Rows, rowToJSON(row))
		}
		for _, row := range res.Removals {
			out.Removals = append(out.Removals, rowToJSON(row))
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}
	printPlanGrouped(w, rd, res)
	return nil
}

func planHasRefused(res plan.Result) bool {
	for _, row := range res.Rows {
		if row.Action == plan.ActionRefused {
			return true
		}
	}
	return false
}

func printPlanGrouped(w io.Writer, rd *redact.Redactor, res plan.Result) {
	groups := map[plan.Action][]plan.Row{}
	removalKeys := map[string]bool{}
	for _, r := range res.Removals {
		removalKeys[r.Namespace+"/"+r.Name] = true
	}
	for _, row := range res.Rows {
		if removalKeys[row.Namespace+"/"+row.Name] {
			continue
		}
		groups[row.Action] = append(groups[row.Action], row)
	}

	printGroup := func(title string, action plan.Action) {
		rows := groups[action]
		if len(rows) == 0 {
			return
		}
		fmt.Fprint(w, MaskedPrint(rd, fmt.Sprintf("\n%s:\n", title)))
		for _, row := range rows {
			printPlanRow(w, rd, row)
		}
	}

	printGroup("Stuck", plan.ActionStuck)
	printGroup("Refused", plan.ActionRefused)
	printGroup("Installs", plan.ActionInstall)
	printGroup("Upgrades", plan.ActionUpgrade)
	printGroup("Adopts", plan.ActionAdopt)
	printGroup("Unchanged", plan.ActionUnchanged)

	if len(res.Removals) > 0 {
		fmt.Fprint(w, MaskedPrint(rd, "\nRemovals:\n"))
		for _, row := range res.Removals {
			printPlanRow(w, rd, row)
		}
	}

	s := summarizePlan(res)
	parts := []string{}
	if s.Install > 0 {
		parts = append(parts, fmt.Sprintf("%d install", s.Install))
	}
	if s.Upgrade > 0 {
		parts = append(parts, fmt.Sprintf("%d upgrade", s.Upgrade))
	}
	if s.Adopt > 0 {
		parts = append(parts, fmt.Sprintf("%d adopt", s.Adopt))
	}
	if s.Remove > 0 {
		parts = append(parts, fmt.Sprintf("%d removal", s.Remove))
	}
	if s.Refused > 0 {
		parts = append(parts, fmt.Sprintf("%d refused", s.Refused))
	}
	if s.Stuck > 0 {
		parts = append(parts, fmt.Sprintf("%d stuck", s.Stuck))
	}
	if s.Unchanged > 0 {
		parts = append(parts, fmt.Sprintf("%d unchanged", s.Unchanged))
	}
	if len(parts) == 0 {
		fmt.Fprint(w, MaskedPrint(rd, "\nPlan: no changes\n"))
		return
	}
	fmt.Fprint(w, MaskedPrint(rd, fmt.Sprintf("\nPlan: %s\n", strings.Join(parts, ", "))))
}

func printPlanRow(w io.Writer, rd *redact.Redactor, row plan.Row) {
	note := ""
	if row.Note != "" {
		note = "  [" + row.Note + "]"
	}
	var line string
	switch row.Action {
	case plan.ActionUnchanged:
		line = fmt.Sprintf("  %s/%s: unchanged (%s)\n", row.Namespace, row.Name, row.ToVersion)
	case plan.ActionInstall:
		line = fmt.Sprintf("  %s/%s: install %s%s\n", row.Namespace, row.Name, row.ToVersion, note)
	case plan.ActionUpgrade:
		line = fmt.Sprintf("  %s/%s: upgrade %s -> %s%s\n", row.Namespace, row.Name, row.FromVersion, row.ToVersion, note)
	case plan.ActionAdopt:
		line = fmt.Sprintf("  %s/%s: adopt %s%s\n", row.Namespace, row.Name, row.ToVersion, note)
	case plan.ActionRemove:
		line = fmt.Sprintf("  %s/%s: remove%s\n", row.Namespace, row.Name, note)
	case plan.ActionRefused, plan.ActionStuck:
		line = fmt.Sprintf("  %s/%s: %s %s\n", row.Namespace, row.Name, row.Action, row.Note)
	default:
		line = fmt.Sprintf("  %s/%s: %s%s\n", row.Namespace, row.Name, row.Action, note)
	}
	fmt.Fprint(w, MaskedPrint(rd, line))
}
