package cluster

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// LabelNamespaceCreatedBy marks namespaces created by the deploy engine.
	LabelNamespaceCreatedBy      = "rtifact.ai/created-by"
	LabelNamespaceCreatedByValue = "deploy"
)

// EnsureNamespace creates name with labels plus the created-by marker, or merges
// labels into an existing namespace (the marker is never added to existing ones).
func EnsureNamespace(ctx context.Context, cs Client, name string, labels map[string]string) error {
	existing, err := cs.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		merged := mergeLabels(existing.Labels, labels)
		if labelsEqual(merged, existing.Labels) {
			return nil
		}
		existing.Labels = merged
		_, err = cs.CoreV1().Namespaces().Update(ctx, existing, metav1.UpdateOptions{})
		return err
	}
	if !apierrors.IsNotFound(err) {
		return err
	}
	createLabels := mergeLabels(nil, labels)
	createLabels[LabelNamespaceCreatedBy] = LabelNamespaceCreatedByValue
	obj := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name, Labels: createLabels}}
	_, err = cs.CoreV1().Namespaces().Create(ctx, obj, metav1.CreateOptions{})
	return err
}

// NamespaceCreatedByDeploy reports whether ns exists and carries the deploy creation marker.
func NamespaceCreatedByDeploy(ctx context.Context, cs Client, ns string) (bool, error) {
	obj, err := cs.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return obj.Labels[LabelNamespaceCreatedBy] == LabelNamespaceCreatedByValue, nil
}

// DeleteNamespace deletes ns; a missing namespace is not an error.
func DeleteNamespace(ctx context.Context, cs Client, ns string) error {
	err := cs.CoreV1().Namespaces().Delete(ctx, ns, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

func mergeLabels(base, extra map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

func labelsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
