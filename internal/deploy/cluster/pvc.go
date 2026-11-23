package cluster

import (
	"context"
	"io"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListPVCs returns PVC names in ns (missing namespace → none).
func ListPVCs(ctx context.Context, cs Client, ns string) ([]string, error) {
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

// DeletePVC deletes one PVC; a missing PVC is not an error.
func DeletePVC(ctx context.Context, cs Client, ns, name string) error {
	err := cs.CoreV1().PersistentVolumeClaims(ns).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

// PVCNamesFromManifest returns PVC names a release owns: PersistentVolumeClaim
// objects plus StatefulSet volumeClaimTemplates names, sorted and de-duplicated.
func PVCNamesFromManifest(manifest string) ([]string, error) {
	if strings.TrimSpace(manifest) == "" {
		return nil, nil
	}
	dec := yaml.NewDecoder(strings.NewReader(manifest))
	seen := map[string]bool{}
	for {
		var doc map[string]any
		err := dec.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		for _, n := range pvcNamesInObject(doc) {
			seen[n] = true
		}
	}
	out := make([]string, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Strings(out)
	return out, nil
}

func pvcNamesInObject(doc map[string]any) []string {
	if doc == nil {
		return nil
	}
	switch kind, _ := doc["kind"].(string); kind {
	case "PersistentVolumeClaim":
		if n := metadataName(doc); n != "" {
			return []string{n}
		}
	case "StatefulSet":
		spec, _ := doc["spec"].(map[string]any)
		vcts, _ := spec["volumeClaimTemplates"].([]any)
		var out []string
		for _, v := range vcts {
			if m, ok := v.(map[string]any); ok {
				if n := metadataName(m); n != "" {
					out = append(out, n)
				}
			}
		}
		return out
	}
	return nil
}

func metadataName(obj map[string]any) string {
	md, _ := obj["metadata"].(map[string]any)
	n, _ := md["name"].(string)
	return n
}
