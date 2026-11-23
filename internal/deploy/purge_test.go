package deploy

import (
	"bytes"
	"context"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/cluster"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const pvcManifestA = `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data-a
`

func pvc(ns, name string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name}}
}

func TestPurge_removingOneReleaseKeepsSiblingPVC(t *testing.T) {
	ctx := context.Background()
	cs := fake.NewSimpleClientset(pvc("shared", "data-a"), pvc("shared", "data-b"))
	if err := cluster.EnsureNamespace(ctx, cs, "shared", nil); err != nil {
		t.Fatal(err)
	}
	remaining := map[string]inventory.Entry{
		entryKey("b", "shared"): {Name: "b", Namespace: "shared"},
	}
	opt := purgeRemovalOpts{ReleaseName: "a", ReleaseNS: "shared", Manifest: pvcManifestA}
	if err := reportOrPurgeRemovalData(ctx, cs, &Document{}, remaining, opt, true, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	left, err := cluster.ListPVCs(ctx, cs, "shared")
	if err != nil {
		t.Fatal(err)
	}
	if len(left) != 1 || left[0] != "data-b" {
		t.Fatalf("PVCs left: %v (want only data-b)", left)
	}
	if _, err := cs.CoreV1().Namespaces().Get(ctx, "shared", metav1.GetOptions{}); err != nil {
		t.Fatalf("namespace still used by b must remain: %v", err)
	}
}

func TestPurge_namespaceNotCreatedByDeployIsKept(t *testing.T) {
	ctx := context.Background()
	preexisting := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "team"}}
	cs := fake.NewSimpleClientset(preexisting)
	opt := purgeRemovalOpts{ReleaseName: "a", ReleaseNS: "team"}
	if err := reportOrPurgeRemovalData(ctx, cs, &Document{}, nil, opt, true, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if _, err := cs.CoreV1().Namespaces().Get(ctx, "team", metav1.GetOptions{}); err != nil {
		t.Fatalf("pre-existing namespace must not be purged: %v", err)
	}
}

func TestPurge_deployCreatedEmptyNamespaceIsDeleted(t *testing.T) {
	ctx := context.Background()
	cs := fake.NewSimpleClientset()
	if err := cluster.EnsureNamespace(ctx, cs, "mine", nil); err != nil {
		t.Fatal(err)
	}
	opt := purgeRemovalOpts{ReleaseName: "a", ReleaseNS: "mine"}
	if err := reportOrPurgeRemovalData(ctx, cs, &Document{}, nil, opt, true, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if _, err := cs.CoreV1().Namespaces().Get(ctx, "mine", metav1.GetOptions{}); err == nil {
		t.Fatal("deploy-created unused namespace should be deleted on purge")
	}
}
