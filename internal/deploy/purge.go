package deploy

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
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

// reportOrPurgeRemovalData lists retained PVCs/namespaces or deletes them when purge is set.
func reportOrPurgeRemovalData(ctx context.Context, kcfg kube.Config, doc *Document,
	remaining map[string]inventory.Entry, ns string, purge bool, w io.Writer) error {
	pvcs, err := listPVCs(ctx, kcfg, ns)
	if err != nil {
		return err
	}
	nsRemovable := !releaseStillUsesNamespace(doc, remaining, ns)
	if !purge {
		if len(pvcs) > 0 {
			fmt.Fprintf(w, "    kept PVCs in %s: %s\n", ns, strings.Join(pvcs, ", "))
		}
		if nsRemovable {
			fmt.Fprintf(w, "    kept namespace: %s\n", ns)
		}
		return nil
	}
	cs, err := kube.Clientset(kcfg)
	if err != nil {
		return err
	}
	for _, name := range pvcs {
		fmt.Fprintf(w, "    purge: deleting PVC %s/%s\n", ns, name)
		if err := cs.CoreV1().PersistentVolumeClaims(ns).Delete(ctx, name, metav1.DeleteOptions{}); err != nil &&
			!apierrors.IsNotFound(err) {
			return err
		}
	}
	if nsRemovable {
		fmt.Fprintf(w, "    purge: deleting namespace %s\n", ns)
		err := cs.CoreV1().Namespaces().Delete(ctx, ns, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}