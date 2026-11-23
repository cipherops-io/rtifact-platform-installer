package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
)

// StatusSchemaVersion is the `status -o json` document version.
const StatusSchemaVersion = 1

// Release states reported by `status`.
const (
	StateDeployed  = "deployed"  // live version matches the file and Helm reports deployed
	StateDrifted   = "drifted"   // live version differs from the file
	StateMissing   = "missing"   // in the file, not installed
	StatePending   = "pending"   // Helm operation in progress or interrupted (pending-*, uninstalling)
	StateFailed    = "failed"    // last Helm operation failed
	StateUnmanaged = "unmanaged" // installed but not labelled for this deployment
	StateOrphaned  = "orphaned"  // in inventory, no longer in the file
)

type statusRow struct {
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
	Chart         string `json:"chart"`
	Wanted        string `json:"wanted,omitempty"`
	Deployed      string `json:"deployed,omitempty"`
	Status        string `json:"status,omitempty"`
	State         string `json:"state"`
	InventoryOnly bool   `json:"inventory_only,omitempty"`
}

type statusJSON struct {
	SchemaVersion int         `json:"schema_version"`
	Deployment    string      `json:"deployment"`
	Releases      []statusRow `json:"releases"`
}

// Status prints per-release wanted vs live versions.
func Status(ctx context.Context, o RunOptions) error {
	sess, err := OpenSession(ctx, o)
	if err != nil {
		return err
	}
	doc := sess.Doc
	if !o.JSONOutput {
		PrintBanner(o.err(), doc, sess.Disc, sess.Kube, sess.Identity)
	}
	deployment := sess.DeploymentID()
	cs, err := sess.helm(o, nil)
	if err != nil {
		return err
	}
	inv, err := inventory.Read(ctx, sess.Client, deployment)
	if err != nil {
		return err
	}
	rows, err := buildStatusRows(ctx, cs, doc, inv)
	if err != nil {
		return err
	}
	if o.JSONOutput {
		enc := json.NewEncoder(o.out())
		enc.SetIndent("", "  ")
		return enc.Encode(statusJSON{SchemaVersion: StatusSchemaVersion, Deployment: deployment, Releases: rows})
	}
	for _, r := range rows {
		fmt.Fprintf(o.out(), "%s/%s  state=%s  chart=%s  wanted=%s  deployed=%s  status=%s\n",
			r.Namespace, r.Name, r.State, r.Chart, dash(r.Wanted), dash(r.Deployed), dash(r.Status))
	}
	return nil
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func rowState(r statusRow, live *plan.Observed) string {
	switch {
	case r.InventoryOnly:
		return StateOrphaned
	case live == nil:
		return StateMissing
	case !live.ManagedByDeployment:
		return StateUnmanaged
	case live.Status == plan.StatusFailed:
		return StateFailed
	case live.Status != plan.StatusDeployed:
		return StatePending
	case live.Version != r.Wanted:
		return StateDrifted
	default:
		return StateDeployed
	}
}

func statusObserveTargets(doc *Document, inv []inventory.Entry) []plan.Wanted {
	want := make([]plan.Wanted, 0, len(doc.Releases))
	for _, r := range doc.Releases {
		want = append(want, plan.Wanted{Name: r.Name, Namespace: r.Namespace})
	}
	return observeTargets(want, inv)
}

func mergeStatusRows(doc *Document, inv []inventory.Entry, live []plan.Observed) []statusRow {
	inFile := map[string]Release{}
	for _, r := range doc.Releases {
		inFile[entryKey(r.Name, r.Namespace)] = r
	}
	liveBy := map[string]plan.Observed{}
	for _, l := range live {
		liveBy[entryKey(l.Name, l.Namespace)] = l
	}
	var rows []statusRow
	for _, r := range doc.Releases {
		k := entryKey(r.Name, r.Namespace)
		row := statusRow{Name: r.Name, Namespace: r.Namespace, Chart: r.Chart, Wanted: r.Version}
		var live *plan.Observed
		if l, ok := liveBy[k]; ok {
			row.Deployed, row.Status, live = l.Version, l.Status, &l
		}
		row.State = rowState(row, live)
		rows = append(rows, row)
	}
	for _, e := range inv {
		k := entryKey(e.Name, e.Namespace)
		if _, ok := inFile[k]; ok {
			continue
		}
		row := statusRow{
			Name: e.Name, Namespace: e.Namespace, Chart: e.Chart,
			InventoryOnly: true,
		}
		if l, ok := liveBy[k]; ok {
			row.Deployed, row.Status = l.Version, l.Status
		}
		row.State = rowState(row, nil)
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Namespace != rows[j].Namespace {
			return rows[i].Namespace < rows[j].Namespace
		}
		return rows[i].Name < rows[j].Name
	})
	return rows
}

func buildStatusRows(ctx context.Context, cs helmclient.Interface, doc *Document, inv []inventory.Entry) ([]statusRow, error) {
	live, err := cs.Observe(ctx, helmclient.ObserveOptions{Wanted: statusObserveTargets(doc, inv)})
	if err != nil {
		return nil, err
	}
	return mergeStatusRows(doc, inv, live), nil
}
