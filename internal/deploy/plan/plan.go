// Package plan decides what a deploy run will do. It takes plain data in and
// returns decisions, with no cluster, Helm or registry calls, so every guard
// is unit-testable without a cluster.
package plan

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

// Action is the planned change for one release.
type Action string

const (
	ActionInstall   Action = "install"
	ActionUpgrade   Action = "upgrade"
	ActionUnchanged Action = "unchanged"
	ActionRemove    Action = "remove"
	ActionAdopt     Action = "adopt"
	ActionStuck     Action = "stuck"
	ActionRefused   Action = "refused"
)

// Helm release statuses the planner distinguishes.
const (
	StatusDeployed     = "deployed"
	StatusUninstalling = "uninstalling"
)

// Wanted describes desired state from the deploy file.
type Wanted struct {
	Name, Namespace string
	Chart, Version  string
	// ChartDigest is the registry (OCI) digest of the chart at Version.
	// Empty means the caller could not resolve it; the chart digest is then
	// not part of the unchanged comparison.
	ChartDigest string
	// ValuesDigest is the digest of the unresolved values file content.
	ValuesDigest string
	Requires     []string
	Protect      bool
}

// Observed is live cluster state for a release.
type Observed struct {
	Name, Namespace string
	Chart, Version  string
	// ChartDigest and ValuesDigest are optional: set when the cluster itself
	// records them. Otherwise the inventory entry supplies them.
	ChartDigest         string
	ValuesDigest        string
	Status              string // deployed, pending-upgrade, pending-install, failed, etc.
	Revision            int
	ManagedByDeployment bool
}

// Entry is one managed release in inventory (schema v2).
type Entry struct {
	Name         string    `json:"name"`
	Namespace    string    `json:"namespace"`
	Chart        string    `json:"chart"`
	Version      string    `json:"version"`
	ChartDigest  string    `json:"chart_digest,omitempty"`
	ValuesDigest string    `json:"values_digest"`
	Requires     []string  `json:"requires,omitempty"`
	Protect      bool      `json:"protect,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Row is one line in a deploy plan.
type Row struct {
	Name, Namespace string
	Action          Action
	FromVersion     string
	ToVersion       string
	Note            string
}

// Input is pure data for planning (no cluster / Helm calls).
type Input struct {
	Wanted     []Wanted
	Observed   []Observed
	Inventory  []Entry
	Deployment string
	// FilePath is the deploy file, used to print recovery commands.
	FilePath string
}

// Options toggles guard overrides.
type Options struct {
	AllowDestroy     bool
	AllowMassRemoval bool
	AllowDowngrade   bool
	AllowAdopt       bool
	Purge            bool
	Only             string
	OnlyClosure      map[string]bool
}

// Result is the full plan.
type Result struct {
	// Rows holds wanted-release rows followed by removal rows.
	Rows []Row
	// Removals holds removal rows only, ordered so dependents go first.
	Removals []Row
	// HasChange is true when at least one row will change the cluster or
	// needs attention. Refused and unchanged rows do not count.
	HasChange bool
}

// Build computes install/upgrade/unchanged/remove decisions.
func Build(in Input, opts Options) Result {
	obsByKey := map[string]Observed{}
	for _, o := range in.Observed {
		obsByKey[key(o.Name, o.Namespace)] = o
	}
	invByKey := map[string]Entry{}
	for _, e := range in.Inventory {
		invByKey[key(e.Name, e.Namespace)] = e
	}

	var res Result
	wantedKeys := map[string]bool{}
	for _, w := range in.Wanted {
		wantedKeys[key(w.Name, w.Namespace)] = true
	}

	for _, w := range in.Wanted {
		if opts.Only != "" && !opts.OnlyClosure[w.Name] {
			continue
		}
		k := key(w.Name, w.Namespace)
		o, live := obsByKey[k]
		e, inInv := invByKey[k]
		var ep *Entry
		if inInv {
			ep = &e
		}
		res.Rows = append(res.Rows, decideWanted(w, o, live, ep, in, opts))
	}
	blockDependents(&res, in)

	res.Removals = decideRemovals(in, opts, wantedKeys, obsByKey)
	res.Rows = append(res.Rows, res.Removals...)

	for _, r := range res.Rows {
		switch r.Action {
		case ActionInstall, ActionUpgrade, ActionAdopt, ActionRemove, ActionStuck:
			res.HasChange = true
		}
	}
	return res
}

func decideWanted(w Wanted, o Observed, live bool, inv *Entry, in Input, opts Options) Row {
	row := Row{Name: w.Name, Namespace: w.Namespace, ToVersion: w.Version}
	wv, err := semver.NewVersion(w.Version)
	if err != nil {
		row.Action = ActionRefused
		row.Note = fmt.Sprintf("wanted version %q is not valid semver", w.Version)
		return row
	}
	if !live {
		row.Action = ActionInstall
		if inv != nil {
			row.Note = "in inventory but not found in the cluster; reinstalling"
		}
		return row
	}
	row.FromVersion = o.Version

	if isStuck(o.Status) {
		row.Action = ActionStuck
		row.ToVersion = ""
		row.Note = fmt.Sprintf("release is %s; recover with: %s", o.Status, RollbackCommand(in.FilePath, w.Name))
		return row
	}
	if o.Status == StatusUninstalling {
		row.Action = ActionRefused
		row.Note = "release is being uninstalled; wait for it to finish and re-run"
		return row
	}
	if !o.ManagedByDeployment {
		if opts.AllowAdopt {
			row.Action = ActionAdopt
		} else {
			row.Action = ActionRefused
			row.Note = "release exists but is not managed by this deployment (use --adopt)"
		}
		return row
	}

	ov, err := semver.NewVersion(o.Version)
	if err != nil {
		row.Action = ActionRefused
		row.Note = fmt.Sprintf("deployed version %q is not valid semver", o.Version)
		return row
	}
	if wv.LessThan(ov) && !opts.AllowDowngrade {
		row.Action = ActionRefused
		row.Note = fmt.Sprintf("downgrade %s -> %s blocked (use --allow-downgrade)", o.Version, w.Version)
		return row
	}

	appliedChartDigest, appliedValuesDigest := appliedDigests(o, inv)
	switch {
	case o.Status != StatusDeployed:
		row.Action = ActionUpgrade
		row.Note = fmt.Sprintf("release status is %q, not deployed; re-applying", o.Status)
	case o.Chart != w.Chart:
		row.Action = ActionUpgrade
		row.Note = fmt.Sprintf("chart changed %s -> %s", o.Chart, w.Chart)
	case !wv.Equal(ov):
		row.Action = ActionUpgrade
	case w.ChartDigest != "" && w.ChartDigest != appliedChartDigest:
		row.Action = ActionUpgrade
		row.Note = "chart digest changed for the same version"
	case w.ValuesDigest != appliedValuesDigest:
		row.Action = ActionUpgrade
		row.Note = "values changed"
	default:
		row.Action = ActionUnchanged
	}
	return row
}

// appliedDigests returns the chart and values digests last applied for the
// observed release. Cluster-recorded digests win; the inventory entry is used
// only when it describes the same chart and version that is live.
func appliedDigests(o Observed, inv *Entry) (chartDigest, valuesDigest string) {
	chartDigest, valuesDigest = o.ChartDigest, o.ValuesDigest
	if inv != nil && inv.Chart == o.Chart && sameVersion(inv.Version, o.Version) {
		if chartDigest == "" {
			chartDigest = inv.ChartDigest
		}
		if valuesDigest == "" {
			valuesDigest = inv.ValuesDigest
		}
	}
	return chartDigest, valuesDigest
}

func sameVersion(a, b string) bool {
	av, err1 := semver.NewVersion(a)
	bv, err2 := semver.NewVersion(b)
	if err1 != nil || err2 != nil {
		return false
	}
	return av.Equal(bv)
}

// blockDependents refuses install/upgrade/adopt rows whose required release
// is stuck (or itself blocked), so nothing is applied on top of a pending
// release.
func blockDependents(res *Result, in Input) {
	requires := map[string][]string{}
	for _, w := range in.Wanted {
		requires[w.Name] = w.Requires
	}
	blocked := map[string]bool{}
	for _, r := range res.Rows {
		if r.Action == ActionStuck {
			blocked[r.Name] = true
		}
	}
	for changed := true; changed; {
		changed = false
		for i := range res.Rows {
			r := &res.Rows[i]
			switch r.Action {
			case ActionInstall, ActionUpgrade, ActionAdopt:
			default:
				continue
			}
			for _, req := range requires[r.Name] {
				if !blocked[req] {
					continue
				}
				r.Action = ActionRefused
				r.Note = fmt.Sprintf("blocked: requires %s, which is stuck or blocked; recover with: %s",
					req, RollbackCommand(in.FilePath, req))
				blocked[r.Name] = true
				changed = true
				break
			}
		}
	}
}

// decideRemovals applies the six removal guards. Candidates are inventory
// entries that are absent from the deploy file (guard 1); the cluster is never
// scanned for unrelated releases. Guard 6 (data stays) is a note here and is
// enforced by the uninstall path.
func decideRemovals(in Input, opts Options, wantedKeys map[string]bool, obsByKey map[string]Observed) []Row {
	dependents := map[string][]string{}
	for _, w := range in.Wanted {
		for _, req := range w.Requires {
			dependents[req] = append(dependents[req], w.Name)
		}
	}

	var cands []Entry
	for _, e := range in.Inventory {
		if wantedKeys[key(e.Name, e.Namespace)] {
			continue
		}
		if opts.Only != "" && !opts.OnlyClosure[e.Name] {
			continue
		}
		cands = append(cands, e)
	}
	cands = RemovalOrder(cands)

	rows := make([]Row, 0, len(cands))
	for _, e := range cands {
		row := Row{Name: e.Name, Namespace: e.Namespace, Action: ActionRemove, FromVersion: e.Version}
		o, live := obsByKey[key(e.Name, e.Namespace)]
		switch {
		case live && !o.ManagedByDeployment:
			row.Action = ActionRefused
			row.Note = "release exists but is not managed by this deployment; not removing"
		case e.Protect && !opts.AllowDestroy:
			row.Action = ActionRefused
			row.Note = "protected release (use --allow-destroy)"
		case len(dependents[e.Name]) > 0:
			row.Action = ActionRefused
			row.Note = "required by " + strings.Join(dependents[e.Name], ", ") + " in the deploy file"
		case !live:
			row.Note = "inventory-only: release not found in the cluster; inventory entry will be dropped"
		default:
			row.Note = dataNote(opts.Purge)
		}
		rows = append(rows, row)
	}

	// Guard 5: more than half the managed releases in one run.
	if len(cands)*2 > len(in.Inventory) && !opts.AllowMassRemoval {
		for i := range rows {
			if rows[i].Action == ActionRemove {
				rows[i].Action = ActionRefused
				rows[i].Note = fmt.Sprintf("mass removal blocked: %d of %d managed releases (use --allow-mass-removal)",
					len(cands), len(in.Inventory))
			}
		}
	}
	return rows
}

func dataNote(purge bool) string {
	if purge {
		return "--purge: release PVCs and deploy-created namespaces may be deleted"
	}
	return "release PVCs and namespaces are kept (use --purge to delete release PVCs)"
}

// RemovalOrder orders inventory entries for uninstall (dependents before requirements).
func RemovalOrder(entries []Entry) []Entry {
	return orderDependentsFirst(entries)
}

// orderDependentsFirst orders candidates so that a release comes before the
// releases it requires. Ties and cycles fall back to name order.
func orderDependentsFirst(cands []Entry) []Entry {
	sorted := append([]Entry(nil), cands...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Name != sorted[j].Name {
			return sorted[i].Name < sorted[j].Name
		}
		return sorted[i].Namespace < sorted[j].Namespace
	})
	remaining := sorted
	var out []Entry
	for len(remaining) > 0 {
		pick := -1
		for i, e := range remaining {
			required := false
			for j, other := range remaining {
				if i == j {
					continue
				}
				for _, req := range other.Requires {
					if req == e.Name {
						required = true
					}
				}
			}
			if !required {
				pick = i
				break
			}
		}
		if pick < 0 { // cycle
			pick = 0
		}
		out = append(out, remaining[pick])
		remaining = append(append([]Entry(nil), remaining[:pick]...), remaining[pick+1:]...)
	}
	return out
}

// RollbackCommand returns the full command that rolls a release back.
// An empty path prints a placeholder.
func RollbackCommand(file, release string) string {
	if file == "" {
		file = "<deploy-file>"
	} else if strings.ContainsAny(file, " \t'\"") {
		file = "'" + strings.ReplaceAll(file, "'", `'\''`) + "'"
	}
	return fmt.Sprintf("rtifact deploy rollback %s %s", file, release)
}

// ReleaseKey uniquely identifies a release within a deployment (name + namespace).
func ReleaseKey(name, ns string) string { return key(name, ns) }

func key(name, ns string) string { return name + "\x00" + ns }

func isStuck(status string) bool {
	switch status {
	case "pending-install", "pending-upgrade", "pending-rollback":
		return true
	default:
		return false
	}
}
