//go:build e2e

package deploy_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/cluster"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/lock"
	releasecommon "helm.sh/helm/v4/pkg/release/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	e2eCleanupDeploy(t, file, o)
	e2eMustApply(t, o)
	nsA := e2eReleaseNS(t, file, "ok-a")
	nsB := e2eReleaseNS(t, file, "ok-b")
	revB0 := e2eHelmRevision(t, "ok-b", nsB)
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
	e2eIntEq(t, "ok-b revision unchanged", e2eHelmRevision(t, "ok-b", nsB), revB0)
	e2eAssertInventory(t, file, "ok-a", "ok-b")
	if rev := e2eHelmRevision(t, "ok-a", nsA); rev < 1 {
		t.Fatalf("ok-a should remain installed, rev=%d", rev)
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
	e2eCleanupDeploy(t, file, o)
	e2eMustApply(t, o)
	file = writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Namespaces = append(doc.Namespaces, deploy.Namespace{Name: "crd-ns"})
		doc.Releases = append(doc.Releases, deploy.Release{
			Name: "crd", Namespace: "crd-ns", Chart: "crd", Version: "2.0.0", CRDs: "upgrade",
		})
	})
	o.FilePath = file
	e2eMustApply(t, o)
	cs := e2eDynamicClient(t, k8s(t))
	u, err := cs.Resource(schema.GroupVersionResource{
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
	e2eCleanupDeploy(t, file, o)
	e2eMustApply(t, o)
	nsB := e2eReleaseNS(t, file, "ok-b")
	file = writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Releases = doc.Releases[:1]
	})
	o.FilePath = file
	e2eMustApply(t, o)
	e2eAssertInventory(t, file, "ok-a")
	e2eHelmReleaseMissing(t, "ok-b", nsB)
}

func TestE2E_S9_ProtectedRefusal(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Releases[0].Protect = true
	})
	o := e2eOpts(t, file)
	e2eCleanupDeploy(t, file, o)
	e2eMustApply(t, o)
	nsB := e2eReleaseNS(t, file, "ok-b")
	file = writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Releases = doc.Releases[1:]
		doc.Releases[0].Requires = nil
	})
	o.FilePath = file
	err := deploy.Apply(context.Background(), o)
	e2eAssertErrContains(t, err, "protect")
	e2eAssertInventory(t, file, "ok-a", "ok-b")
	if e2eHelmRevision(t, "ok-b", nsB) < 1 {
		t.Fatal("ok-b should remain installed")
	}
}

func TestE2E_S10_MassRemovalGuard(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(*deploy.Document) {})
	o := e2eOpts(t, file)
	e2eCleanupDeploy(t, file, o)
	e2eMustApply(t, o)
	nsA := e2eReleaseNS(t, file, "ok-a")
	file = writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Releases = nil
	})
	o.FilePath = file
	o.DetailedExitCode = true
	_, _, err := e2ePlanExpect(t, o, 1, true)
	if err == nil || !strings.Contains(err.Error(), "mass removal") {
		t.Fatalf("expected mass removal refusal, err=%v", err)
	}
	err = deploy.Apply(context.Background(), o)
	if err == nil {
		t.Fatal("expected apply to refuse mass removal")
	}
	e2eAssertInventory(t, file, "ok-a", "ok-b")
	e2eIntEq(t, "ok-a still installed", e2eHelmRevision(t, "ok-a", nsA), 1)
}

func TestE2E_S12_DeployLock(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(*deploy.Document) {})
	o := e2eOpts(t, file)
	e2eCleanupDeploy(t, file, o)
	ctx := context.Background()
	cs := e2eClientset(t)
	if err := cluster.EnsureNamespace(ctx, cs, inventory.SystemNS, nil); err != nil {
		t.Fatal(err)
	}
	held, err := lock.Acquire(ctx, cs, deploy.DeploymentID(e2eDoc(t, file)), "other-operator@ci 1+held", false)
	if err != nil {
		t.Fatal(err)
	}
	err = deploy.Apply(ctx, o)
	e2eLockHeldInErr(t, err)
	e2eAssertErrContains(t, err, "other-operator@ci")
	e2eHelmReleaseMissing(t, "ok-a", e2eReleaseNS(t, file, "ok-a"))
	if err := held.Release(ctx, cs); err != nil {
		t.Fatal(err)
	}
	e2eMustApply(t, o)
	e2eAssertLockFree(t, file)
}

func TestE2E_S11_StuckReportsRecovery(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(*deploy.Document) {})
	o := e2eOpts(t, file)
	e2eCleanupDeploy(t, file, o)
	e2eMustApply(t, o)
	nsA := e2eReleaseNS(t, file, "ok-a")
	e2eHelmUpgradeNoWait(t, "ok-a", nsA, "broken", "1.0.0")
	e2eMarkLatestRevision(t, "ok-a", nsA, releasecommon.StatusPendingUpgrade)
	var buf strings.Builder
	oo := e2eOpts(t, file)
	oo.Stdout = &buf
	oo.DetailedExitCode = true
	code, err := deploy.Plan(context.Background(), oo)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if code != 2 {
		t.Fatalf("expected plan exit 2 for a stuck release, got %d with:\n%s", code, buf.String())
	}
	e2eAssertPlanStuckRecovery(t, buf.String(), file, "ok-a")
	if err := deploy.Rollback(context.Background(), o, "ok-a"); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	_, status, err := e2eHelmReleaseInfo(t, "ok-a", nsA)
	if err != nil {
		t.Fatal(err)
	}
	if status != "deployed" {
		t.Fatalf("ok-a status after rollback: got %q want deployed", status)
	}
}

func TestE2E_S14_InvalidValuesFile(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad-values.yaml")
	if err := os.WriteFile(bad, []byte("{unterminated"), 0o644); err != nil {
		t.Fatal(err)
	}
	file := writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Releases[0].ValuesFiles = []string{"bad-values.yaml"}
	})
	o := e2eOpts(t, file)
	_, err := deploy.Plan(context.Background(), o)
	if err == nil {
		t.Fatal("expected plan to fail on invalid values file")
	}
	if !strings.Contains(err.Error(), "bad-values.yaml") {
		t.Fatalf("unexpected error: %v", err)
	}
	err = deploy.Apply(context.Background(), o)
	if err == nil {
		t.Fatal("expected apply to fail on invalid values file")
	}
}

func TestE2E_S15_PullSecretRefresh(t *testing.T) {
	dir := t.TempDir()
	secretName := "rtifact-registry-pull"
	file := writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Namespaces[0].PullSecret = true
	})
	o := e2eOpts(t, file)
	e2eCleanupDeploy(t, file, o)
	e2eMustApply(t, o)
	nsA := e2eReleaseNS(t, file, "ok-a")
	cs, err := kube.Clientset(k8s(t))
	if err != nil {
		t.Fatal(err)
	}
	s1, err := cs.CoreV1().Secrets(nsA).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("RTIFACT_REGISTRY_USER", e2eEnv(t, "DEPLOY_E2E_REGISTRY_ROTATED_USER"))
	t.Setenv("RTIFACT_REGISTRY_PASSWORD", e2eEnv(t, "DEPLOY_E2E_REGISTRY_ROTATED_PASSWORD"))
	o.FilePath = file
	e2eMustApply(t, o)
	s2, err := cs.CoreV1().Secrets(nsA).Get(context.Background(), secretName, metav1.GetOptions{})
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
	e2eCleanupDeploy(t, file, o)
	e2eMustApply(t, o)
	nsA := e2eReleaseNS(t, file, "ok-a")
	file = writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Releases[0].Version = "1.0.0"
	})
	o.FilePath = file
	o.DetailedExitCode = true
	_, _, err := e2ePlanExpect(t, o, 1, true)
	if err == nil || !strings.Contains(err.Error(), "downgrade") {
		t.Fatalf("expected downgrade refusal on plan, err=%v", err)
	}
	err = deploy.Apply(context.Background(), o)
	if err == nil {
		t.Fatal("expected downgrade refusal")
	}
	revHigh := e2eHelmRevision(t, "ok-a", nsA)
	o.AllowDowngrade = true
	e2eMustApply(t, o)
	e2eIntEq(t, "ok-a revision after downgrade", e2eHelmRevision(t, "ok-a", nsA), revHigh+1)
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
	e2eCleanupDeploy(t, file, o)
	e2eMustApply(t, o)
	ns := e2eReleaseNS(t, file, "stateful")
	if err := deploy.Destroy(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	e2eHelmReleaseMissing(t, "stateful", ns)
	cs, err := kube.Clientset(k8s(t))
	if err != nil {
		t.Fatal(err)
	}
	_, err = cs.CoreV1().PersistentVolumeClaims(ns).Get(context.Background(), pvcName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("expected PVC to remain after destroy: %v", err)
	}
}