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

type statusRow struct {
	Name, Namespace, Chart, Wanted, Deployed, Status string
	InventoryOnly                                  bool
}

// Status prints per-release wanted vs live versions.
func Status(ctx context.Context, o RunOptions) error {
	sess, err := OpenSession(ctx, o)
	if err != nil {
		return err
	}
	doc, disc, kcfg := sess.Doc, sess.Disc, sess.Kube
	if !o.JSONOutput {
		PrintBanner(o.err(), doc, disc, kcfg, sess.Identity)
	}
	deployment := DeploymentID(doc)
	cs, err := newHelmClient(kcfg, doc, nil, o)
	if err != nil {
		return err
	}
	inv, err := inventory.Read(ctx, kcfg, deployment)
	if err != nil {
		return err
	}
	rows, err := buildStatusRows(ctx, cs, doc, inv)
	if err != nil {
		return err
	}
	if o.JSONOutput {
		return json.NewEncoder(o.out()).Encode(rows)
	}
	for _, r := range rows {
		wanted := r.Wanted
		if r.InventoryOnly && wanted == "" {
			wanted = "-"
		}
		tag := ""
		if r.InventoryOnly {
			tag = "  [inventory-only]"
		}
		fmt.Fprintf(o.out(), "%s/%s  chart=%s  wanted=%s  deployed=%s  status=%s%s\n",
			r.Namespace, r.Name, r.Chart, wanted, r.Deployed, r.Status, tag)
	}
	return nil
}

func statusObserveTargets(doc *Document, inv []inventory.Entry) []plan.Wanted {
	var observe []plan.Wanted
	seen := map[string]bool{}
	for _, r := range doc.Releases {
		k := entryKey(r.Name, r.Namespace)
		if seen[k] {
			continue
		}
		seen[k] = true
		observe = append(observe, plan.Wanted{Name: r.Name, Namespace: r.Namespace})
	}
	for _, e := range inv {
		k := entryKey(e.Name, e.Namespace)
		if seen[k] {
			continue
		}
		seen[k] = true
		observe = append(observe, plan.Wanted{Name: e.Name, Namespace: e.Namespace})
	}
	return observe
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
		if l, ok := liveBy[k]; ok {
			row.Deployed = l.Version
			row.Status = l.Status
		}
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
			row.Deployed = l.Version
			row.Status = l.Status
		}
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

func buildStatusRows(ctx context.Context, cs *helmclient.Client, doc *Document, inv []inventory.Entry) ([]statusRow, error) {
	live, err := cs.Observe(ctx, helmclient.ObserveOptions{Wanted: statusObserveTargets(doc, inv)})
	if err != nil {
		return nil, err
	}
	return mergeStatusRows(doc, inv, live), nil
}
