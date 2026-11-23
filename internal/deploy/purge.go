package deploy

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/cluster"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
	"github.com/cipherops-io/rtifact-platform-installer/internal/engine"
)

func releaseStillUsesNamespace(doc *Document, remaining map[string]inventory.Entry, ns string) bool {
	for _, e := range remaining {
		if e.Namespace == ns {
			return true
		}
	}
	for _, r := range doc.Releases {
		if r.Namespace == ns {
			return true
		}
	}
	return false
}

type purgeRemovalOpts struct {
	ReleaseName string
	ReleaseNS   string
	Manifest    string
}

// purgePreviewLines lists the PVCs and namespaces --purge would delete for removal rows.
func purgePreviewLines(ctx context.Context, cs cluster.Client, doc *Document,
	remaining map[string]inventory.Entry, rows []purgeRemovalOpts) ([]string, error) {
	var lines []string
	nsDelete := map[string]bool{}
	for _, row := range rows {
		targets, err := cluster.PVCNamesFromManifest(row.Manifest)
		if err != nil {
			return nil, err
		}
		for _, pvc := range targets {
			lines = append(lines, fmt.Sprintf("PVC %s/%s (release %s)", row.ReleaseNS, pvc, row.ReleaseName))
		}
		if releaseStillUsesNamespace(doc, remaining, row.ReleaseNS) {
			continue
		}
		ok, err := cluster.NamespaceCreatedByDeploy(ctx, cs, row.ReleaseNS)
		if err != nil {
			return nil, err
		}
		if ok {
			nsDelete[row.ReleaseNS] = true
		}
	}
	var nss []string
	for ns := range nsDelete {
		nss = append(nss, ns)
	}
	sort.Strings(nss)
	for _, ns := range nss {
		lines = append(lines, "namespace "+ns)
	}
	return lines, nil
}

// reportOrPurgeRemovalData lists or deletes the PVCs owned by the removed release's manifest,
// then the namespace when deploy created it and nothing else uses it.
func reportOrPurgeRemovalData(ctx context.Context, cs cluster.Client, doc *Document,
	remaining map[string]inventory.Entry, opt purgeRemovalOpts, purge bool, w io.Writer) error {
	owned, err := livePVCsOwnedBy(ctx, cs, opt)
	if err != nil {
		return err
	}
	if !purge && len(owned) > 0 {
		fmt.Fprintf(w, "    kept PVCs for %s/%s: %s\n", opt.ReleaseNS, opt.ReleaseName, strings.Join(owned, ", "))
	}
	if purge {
		for _, name := range owned {
			fmt.Fprintf(w, "    purge: deleting PVC %s/%s (release %s)\n", opt.ReleaseNS, name, opt.ReleaseName)
			if err := cluster.DeletePVC(ctx, cs, opt.ReleaseNS, name); err != nil {
				return err
			}
		}
	}
	return reportNamespaceRetention(ctx, cs, doc, remaining, opt.ReleaseNS, purge, w)
}

func livePVCsOwnedBy(ctx context.Context, cs cluster.Client, opt purgeRemovalOpts) ([]string, error) {
	targets, err := cluster.PVCNamesFromManifest(opt.Manifest)
	if err != nil || len(targets) == 0 {
		return nil, err
	}
	live, err := cluster.ListPVCs(ctx, cs, opt.ReleaseNS)
	if err != nil {
		return nil, err
	}
	liveSet := map[string]bool{}
	for _, n := range live {
		liveSet[n] = true
	}
	var out []string
	for _, n := range targets {
		if liveSet[n] {
			out = append(out, n)
		}
	}
	return out, nil
}

func reportNamespaceRetention(ctx context.Context, cs cluster.Client, doc *Document,
	remaining map[string]inventory.Entry, ns string, purge bool, w io.Writer) error {
	if releaseStillUsesNamespace(doc, remaining, ns) {
		return nil
	}
	mayDelete, err := cluster.NamespaceCreatedByDeploy(ctx, cs, ns)
	if err != nil {
		return err
	}
	switch {
	case !mayDelete && !purge:
		fmt.Fprintf(w, "    kept namespace: %s (not created by deploy)\n", ns)
		return nil
	case !mayDelete:
		return nil
	case !purge:
		fmt.Fprintf(w, "    kept namespace: %s\n", ns)
		return nil
	}
	fmt.Fprintf(w, "    purge: deleting namespace %s\n", ns)
	return cluster.DeleteNamespace(ctx, cs, ns)
}

// confirmPurgeRemovals prints what --purge will delete and asks a second time.
func confirmPurgeRemovals(ctx context.Context, ar *applyRun, removals []plan.Row) error {
	remaining := map[string]inventory.Entry{}
	for k, e := range ar.entries {
		remaining[k] = e
	}
	for _, row := range removals {
		delete(remaining, entryKey(row.Name, row.Namespace))
	}
	var previews []purgeRemovalOpts
	for _, row := range removals {
		manifest, err := ar.helm.ReleaseManifest(ctx, row.Name, row.Namespace)
		if err != nil {
			return err
		}
		previews = append(previews, purgeRemovalOpts{ReleaseName: row.Name, ReleaseNS: row.Namespace, Manifest: manifest})
	}
	lines, err := purgePreviewLines(ctx, ar.sess.Client, ar.sess.Doc, remaining, previews)
	if err != nil {
		return err
	}
	if len(lines) == 0 {
		fmt.Fprintln(ar.o.out(), "  purge: no release PVCs or deploy-created namespaces to delete")
	} else {
		fmt.Fprintln(ar.o.out(), "  purge would delete:")
		for _, line := range lines {
			fmt.Fprintf(ar.o.out(), "    %s\n", line)
		}
	}
	return engine.ConfirmDeployPurge(ar.o.Yes)
}
