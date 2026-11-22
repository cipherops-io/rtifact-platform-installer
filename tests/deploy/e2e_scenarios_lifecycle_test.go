//go:build e2e

package deploy_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/lock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func k8s(t *testing.T) kube.Config {
	t.Helper()
	return kube.Config{Kubeconfig: e2eKubeconfig(t)}
}

func TestE2E_S5_BrokenRollback(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Defaults.Timeout = "2m"
	})
	o := e2eOpts(t, file)
	if err := deploy.Apply(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	file = writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Defaults.Timeout = "90s"
		okA, okB := doc.Releases[0], doc.Releases[1]
		doc.Releases = []deploy.Release{
			okA,
			{Name: "broken", Namespace: "deploy-a", Chart: "broken", Version: "1.0.0", CRDs: "chart"},
			okB,
		}
	})
	o.FilePath = file
	err := deploy.Apply(context.Background(), o)
	if err == nil {
		t.Fatal("expected broken release failure")
	}
}

func TestE2E_S6_CRDUpgrade(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Namespaces = append(doc.Namespaces, deploy.Namespace{Name: "crd-ns"})
		doc.Releases = append(doc.Releases, deploy.Release{
			Name: "crd", Namespace: "crd-ns", Chart: "crd", Version: "1.0.0", CRDs: "upgrade",
		})
	})
	o := e2eOpts(t, file)
	if err := deploy.Apply(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	file = writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Namespaces = append(doc.Namespaces, deploy.Namespace{Name: "crd-ns"})
		doc.Releases = append(doc.Releases, deploy.Release{
			Name: "crd", Namespace: "crd-ns", Chart: "crd", Version: "2.0.0", CRDs: "upgrade",
		})
	})
	o.FilePath = file
	if err := deploy.Apply(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	cs, err := kube.DynamicClient(k8s(t))
	if err != nil {
		t.Fatal(err)
	}
	u, err := cs.Resource(metav1.GroupVersionResource{
		Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions",
	}).Get(context.Background(), "widgets.deploytests.rtifact.ai", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(u.Object)
	if !strings.Contains(string(b), `"color"`) {
		t.Fatalf("expected color field in upgraded CRD")
	}
}

func TestE2E_S7_RemoveRelease(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(*deploy.Document) {})
	o := e2eOpts(t, file)
	_ = deploy.Apply(context.Background(), o)
	file = writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Releases = doc.Releases[:1]
	})
	o.FilePath = file
	if err := deploy.Apply(context.Background(), o); err != nil {
		t.Fatal(err)
	}
}

func TestE2E_S9_ProtectedRefusal(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Releases[0].Protect = true
	})
	o := e2eOpts(t, file)
	_ = deploy.Apply(context.Background(), o)
	file = writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Releases = doc.Releases[1:]
	})
	o.FilePath = file
	err := deploy.Apply(context.Background(), o)
	if err == nil {
		t.Fatal("expected protected removal refusal")
	}
}

func TestE2E_S10_MassRemovalGuard(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(*deploy.Document) {})
	o := e2eOpts(t, file)
	_ = deploy.Apply(context.Background(), o)
	file = writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Releases = nil
	})
	o.FilePath = file
	code, err := deploy.Plan(context.Background(), o)
	if err != nil || code != 2 {
		t.Fatalf("plan mass removal: code=%d err=%v", code, err)
	}
	err = deploy.Apply(context.Background(), o)
	if err == nil {
		t.Fatal("expected apply to refuse mass removal")
	}
}

func TestE2E_S12_DeployLock(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(*deploy.Document) {})
	o := e2eOpts(t, file)
	if err := deploy.Apply(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	doc, err := deploy.LoadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	kcfg := k8s(t)
	h1, err := lock.Acquire(context.Background(), kcfg, deploy.DeploymentID(doc), "holder-a", false)
	if err != nil {
		t.Fatal(err)
	}
	defer h1.Release(context.Background(), kcfg, deploy.DeploymentID(doc))
	_, err = lock.Acquire(context.Background(), kcfg, deploy.DeploymentID(doc), "holder-b", false)
	if err == nil {
		t.Fatal("expected lock refusal for second holder")
	}
}

func TestE2E_S15_PullSecretRefresh(t *testing.T) {
	dir := t.TempDir()
	secretName := "rtifact-registry-pull"
	file := writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Namespaces[0].PullSecret = true
	})
	o := e2eOpts(t, file)
	if err := deploy.Apply(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	cs, err := kube.Clientset(k8s(t))
	if err != nil {
		t.Fatal(err)
	}
	s1, err := cs.CoreV1().Secrets("deploy-a").Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("RTIFACT_REGISTRY_PASSWORD", "newpass")
	o.FilePath = file
	if err := deploy.Apply(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	s2, err := cs.CoreV1().Secrets("deploy-a").Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if string(s1.Data[".dockerconfigjson"]) == string(s2.Data[".dockerconfigjson"]) {
		t.Fatal("expected pull secret update after password change")
	}
}

func TestE2E_S16_DowngradeGuard(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Releases[0].Version = "1.0.1"
	})
	o := e2eOpts(t, file)
	_ = deploy.Apply(context.Background(), o)
	file = writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Releases[0].Version = "1.0.0"
	})
	o.FilePath = file
	code, err := deploy.Plan(context.Background(), o)
	if err != nil || code != 2 {
		t.Fatalf("plan downgrade: code=%d err=%v", code, err)
	}
	o.AllowDowngrade = true
	if err := deploy.Apply(context.Background(), o); err != nil {
		t.Fatal(err)
	}
}

func TestE2E_S18_DestroyKeepsPVC(t *testing.T) {
	dir := t.TempDir()
	pvcName := "stateful-data"
	file := writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Namespaces = append(doc.Namespaces, deploy.Namespace{Name: "state"})
		doc.Releases = []deploy.Release{
			{Name: "stateful", Namespace: "state", Chart: "stateful", Version: "1.0.0", CRDs: "chart"},
		}
	})
	o := e2eOpts(t, file)
	if err := deploy.Apply(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	if err := deploy.Destroy(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	cs, err := kube.Clientset(k8s(t))
	if err != nil {
		t.Fatal(err)
	}
	_, err = cs.CoreV1().PersistentVolumeClaims("state").Get(context.Background(), pvcName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("expected PVC to remain after destroy: %v", err)
	}
}
