package deploy

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/engine"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func listPVCs(ctx context.Context, kcfg kube.Config, ns string) ([]string, error) {
	cs, err := kube.Clientset(kcfg)
	if err != nil {
		return nil, err
	}
	list, err := cs.CoreV1().PersistentVolumeClaims(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	out := make([]string, 0, len(list.Items))
	for _, pvc := range list.Items {
		out = append(out, pvc.Name)
	}
	return out, nil
}

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

func namespaceMayDelete(ctx context.Context, kcfg kube.Config, ns string) (bool, error) {
	cs, err := kube.Clientset(kcfg)
	if err != nil {
		return false, err
	}
	obj, err := cs.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return namespaceCreatedByDeploy(obj.Labels), nil
}

type purgeRemovalOpts struct {
	ReleaseName string
	ReleaseNS   string
	Manifest    string
}

// purgePreviewLines lists PVCs and namespaces that --purge would delete for removal rows.
func purgePreviewLines(ctx context.Context, kcfg kube.Config, doc *Document,
	remaining map[string]inventory.Entry, rows []purgeRemovalOpts) ([]string, error) {
	var lines []string
	nsDelete := map[string]bool{}
	for _, row := range rows {
		targets, err := pvcNamesFromManifest(row.Manifest)
		if err != nil {
			return nil, err
		}
		for _, pvc := range targets {
			lines = append(lines, fmt.Sprintf("PVC %s/%s (release %s)", row.ReleaseNS, pvc, row.ReleaseName))
		}
		if releaseStillUsesNamespace(doc, remaining, row.ReleaseNS) {
			continue
		}
		ok, err := namespaceMayDelete(ctx, kcfg, row.ReleaseNS)
		if err != nil {
			return nil, err
		}
		if ok {
			nsDelete[row.ReleaseNS] = true
		}
	}
	for ns := range nsDelete {
		lines = append(lines, fmt.Sprintf("namespace %s", ns))
	}
	return lines, nil
}

// reportOrPurgeRemovalData lists or deletes PVCs tied to the removed release manifest.
func reportOrPurgeRemovalData(ctx context.Context, kcfg kube.Config, doc *Document,
	remaining map[string]inventory.Entry, opt purgeRemovalOpts, purge bool, w io.Writer) error {
	targets, err := pvcNamesFromManifest(opt.Manifest)
	if err != nil {
		return err
	}
	if len(targets) == 0 && opt.Manifest == "" {
		// Release already gone; nothing to scope.
		return reportNamespaceRetention(ctx, kcfg, doc, remaining, opt.ReleaseNS, purge, w)
	}
	cs, err := kube.Clientset(kcfg)
	if err != nil {
		return err
	}
	live, err := listPVCs(ctx, kcfg, opt.ReleaseNS)
	if err != nil {
		return err
	}
	liveSet := map[string]bool{}
	for _, n := range live {
		liveSet[n] = true
	}
	var toHandle []string
	for _, name := range targets {
		if liveSet[name] {
			toHandle = append(toHandle, name)
		}
	}
	if !purge {
		if len(toHandle) > 0 {
			fmt.Fprintf(w, "    kept PVCs for %s/%s: %s\n", opt.ReleaseNS, opt.ReleaseName, strings.Join(toHandle, ", "))
		}
		return reportNamespaceRetention(ctx, kcfg, doc, remaining, opt.ReleaseNS, false, w)
	}
	for _, name := range toHandle {
		fmt.Fprintf(w, "    purge: deleting PVC %s/%s (release %s)\n", opt.ReleaseNS, name, opt.ReleaseName)
		if err := cs.CoreV1().PersistentVolumeClaims(opt.ReleaseNS).Delete(ctx, name, metav1.DeleteOptions{}); err != nil &&
			!apierrors.IsNotFound(err) {
			return err
		}
	}
	return reportNamespaceRetention(ctx, kcfg, doc, remaining, opt.ReleaseNS, purge, w)
}

func reportNamespaceRetention(ctx context.Context, kcfg kube.Config, doc *Document,
	remaining map[string]inventory.Entry, ns string, purge bool, w io.Writer) error {
	if releaseStillUsesNamespace(doc, remaining, ns) {
		return nil
	}
	mayDelete, err := namespaceMayDelete(ctx, kcfg, ns)
	if err != nil {
		return err
	}
	if !mayDelete {
		if !purge {
			fmt.Fprintf(w, "    kept namespace: %s (not created by deploy)\n", ns)
		}
		return nil
	}
	if !purge {
		fmt.Fprintf(w, "    kept namespace: %s\n", ns)
		return nil
	}
	cs, err := kube.Clientset(kcfg)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "    purge: deleting namespace %s\n", ns)
	err = cs.CoreV1().Namespaces().Delete(ctx, ns, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

func fetchReleaseManifest(ctx context.Context, hc *helmclient.Client, name, ns string) (string, error) {
	if hc == nil {
		return "", nil
	}
	return hc.ReleaseManifest(ctx, name, ns)
}

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
		ns := row.Namespace
		manifest, err := fetchReleaseManifest(ctx, ar.helm, row.Name, ns)
		if err != nil {
			return err
		}
		previews = append(previews, purgeRemovalOpts{
			ReleaseName: row.Name, ReleaseNS: ns, Manifest: manifest,
		})
	}
	lines, err := purgePreviewLines(ctx, ar.kcfg, ar.doc, remaining, previews)
	if err != nil {
		return err
	}
	if len(lines) == 0 {
		fmt.Fprintln(ar.o.out(), "  purge: no release PVCs or deploy-created namespaces to delete")
		return engine.ConfirmDeployPurge(ar.o.Yes)
	}
	fmt.Fprintln(ar.o.out(), "  purge would delete:")
	for _, line := range lines {
		fmt.Fprintf(ar.o.out(), "    %s\n", line)
	}
	return engine.ConfirmDeployPurge(ar.o.Yes)
}