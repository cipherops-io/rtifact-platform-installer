//go:build e2e

package deploy_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/cluster"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/lock"
	releasecommon "helm.sh/helm/v4/pkg/release/common"
	release "helm.sh/helm/v4/pkg/release/v1"
	"helm.sh/helm/v4/pkg/storage/driver"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

var safeTestName = regexp.MustCompile(`[^a-z0-9-]+`)

func deploymentPrefix(t *testing.T) string {
	t.Helper()
	name := strings.ToLower(t.Name())
	name = safeTestName.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	if len(name) > 36 {
		name = name[:36]
	}
	return "e2e-" + name
}

func isolateFixture(t *testing.T, doc *deploy.Document) {
	t.Helper()
	p := deploymentPrefix(t)
	doc.NamePrefix = p
	// DeploymentID is customer.code-environment; a per-test environment gives each
	// test its own inventory ConfigMap and lock Lease.
	doc.Environment = p
	repl := map[string]string{
		"deploy-a": p + "-ns-a",
		"deploy-b": p + "-ns-b",
		"crd-ns":   p + "-crd",
		"state":    p + "-state",
	}
	for i := range doc.Namespaces {
		if n, ok := repl[doc.Namespaces[i].Name]; ok {
			doc.Namespaces[i].Name = n
		}
	}
	for i := range doc.Releases {
		if n, ok := repl[doc.Releases[i].Namespace]; ok {
			doc.Releases[i].Namespace = n
		}
	}
}

func e2eCleanupDeploy(t *testing.T, file string, o deploy.RunOptions) {
	t.Helper()
	t.Cleanup(func() {
		co := o
		co.AllowDestroy = true
		co.Yes = true
		if err := deploy.Destroy(context.Background(), co); err != nil {
			t.Errorf("cleanup destroy: %v", err)
		}
	})
}

func e2eMustApply(t *testing.T, o deploy.RunOptions) {
	t.Helper()
	if err := deploy.Apply(context.Background(), o); err != nil {
		t.Fatalf("apply: %v", err)
	}
}

func e2eDoc(t *testing.T, file string) *deploy.Document {
	t.Helper()
	doc, err := deploy.LoadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func e2eReleaseNS(t *testing.T, file, name string) string {
	t.Helper()
	for _, r := range e2eDoc(t, file).Releases {
		if r.Name == name {
			return r.Namespace
		}
	}
	t.Fatalf("release %q not in %s", name, file)
	return ""
}

func e2eK8sCfg() kube.Config {
	return kube.Config{Kubeconfig: os.Getenv("DEPLOY_E2E_KUBECONFIG")}
}

func e2eClientset(t *testing.T) cluster.Client {
	t.Helper()
	cs, err := cluster.New(e2eK8sCfg())
	if err != nil {
		t.Fatal(err)
	}
	return cs
}

// e2eAssertLockFree fails if the deploy Lease for file's deployment still exists.
func e2eAssertLockFree(t *testing.T, file string) {
	t.Helper()
	name := lock.LeaseName(deploy.DeploymentID(e2eDoc(t, file)))
	l, err := e2eClientset(t).CoordinationV1().Leases(inventory.SystemNS).Get(context.Background(), name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Fatalf("deploy lock %s still held by %q", name, *l.Spec.HolderIdentity)
}

func e2eInventory(t *testing.T, file string) []inventory.Entry {
	t.Helper()
	doc := e2eDoc(t, file)
	inv, err := inventory.Read(context.Background(), e2eClientset(t), deploy.DeploymentID(doc))
	if err != nil {
		t.Fatal(err)
	}
	return inv
}

func e2eAssertInventory(t *testing.T, file string, names ...string) {
	t.Helper()
	inv := e2eInventory(t, file)
	if len(inv) != len(names) {
		t.Fatalf("inventory len=%d want %d: %#v", len(inv), len(names), inv)
	}
	want := map[string]bool{}
	for _, n := range names {
		want[n] = true
	}
	for _, e := range inv {
		if !want[e.Name] {
			t.Fatalf("unexpected inventory entry %#v", e)
		}
		delete(want, e.Name)
	}
	if len(want) != 0 {
		t.Fatalf("missing inventory releases: %v", want)
	}
}

func e2eHelmClient(t *testing.T) *helmclient.Client {
	t.Helper()
	cs, err := helmclient.New(helmclient.Options{Kube: e2eK8sCfg()})
	if err != nil {
		t.Fatal(err)
	}
	return cs
}

func e2eHelmRevision(t *testing.T, release, ns string) int {
	t.Helper()
	rev, _, err := e2eHelmReleaseInfo(t, release, ns)
	if err != nil {
		t.Fatal(err)
	}
	return rev
}

func e2eHelmReleaseInfo(t *testing.T, release, ns string) (revision int, status string, err error) {
	t.Helper()
	rel, err := e2eHelmClient(t).Lookup(context.Background(), release, ns)
	if err != nil {
		return 0, "", err
	}
	return rel.Revision, rel.Status, nil
}

func e2eHelmReleaseMissing(t *testing.T, release, ns string) {
	t.Helper()
	_, err := e2eHelmClient(t).Lookup(context.Background(), release, ns)
	if err == nil {
		t.Fatalf("expected helm release %s/%s to be absent", ns, release)
	}
}

// e2ePlanExpect runs Plan; expected refusal errors do not fail the test when allowErr is true.
func e2ePlanExpect(t *testing.T, o deploy.RunOptions, wantCode int, allowErr bool) (code int, stdout string, err error) {
	t.Helper()
	var buf bytes.Buffer
	oo := o
	if oo.Stdout == nil {
		oo.Stdout = &buf
	} else {
		oo.Stdout = ioMultiWriter{o.Stdout, &buf}
	}
	code, err = deploy.Plan(context.Background(), oo)
	if err != nil && !allowErr {
		t.Fatalf("plan: code=%d err=%v\n%s", code, err, buf.String())
	}
	if code != wantCode {
		t.Fatalf("plan exit: got %d want %d (err=%v)\n%s", code, wantCode, err, buf.String())
	}
	return code, buf.String(), err
}

func e2ePlan(t *testing.T, o deploy.RunOptions) (code int, stdout string) {
	t.Helper()
	var buf bytes.Buffer
	oo := o
	if oo.Stdout == nil {
		oo.Stdout = &buf
	} else {
		oo.Stdout = ioMultiWriter{o.Stdout, &buf}
	}
	code, err := deploy.Plan(context.Background(), oo)
	if err != nil {
		t.Fatalf("plan: code=%d err=%v\n%s", code, err, buf.String())
	}
	return code, buf.String()
}

type ioMultiWriter struct {
	a, b interface{ Write([]byte) (int, error) }
}

func (m ioMultiWriter) Write(p []byte) (int, error) {
	if m.a != nil {
		_, _ = m.a.Write(p)
	}
	return m.b.Write(p)
}

// e2eMarkLatestRevision rewrites the newest Helm release record's status, as a
// Helm process killed mid-operation leaves it.
func e2eMarkLatestRevision(t *testing.T, name, ns string, status releasecommon.Status) {
	t.Helper()
	store := driver.NewSecrets(e2eClientset(t).CoreV1().Secrets(ns))
	rels, err := store.Query(map[string]string{"name": name, "owner": "helm"})
	if err != nil {
		t.Fatal(err)
	}
	var latest *release.Release
	for _, rl := range rels {
		if r, ok := rl.(*release.Release); ok && (latest == nil || r.Version > latest.Version) {
			latest = r
		}
	}
	if latest == nil {
		t.Fatalf("no helm release records for %s/%s", ns, name)
	}
	latest.Info.Status = status
	if err := store.Update(fmt.Sprintf("sh.helm.release.v1.%s.v%d", name, latest.Version), latest); err != nil {
		t.Fatal(err)
	}
}

func e2eEnv(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Fatalf("%s not set", key)
	}
	return v
}

func e2eChartRef(t *testing.T) string {
	t.Helper()
	return "oci://" + e2eEnv(t, "DEPLOY_E2E_REGISTRY_HOST") + "/charts"
}

func e2eHelmUpgradeNoWait(t *testing.T, release, ns, chart, version string) {
	t.Helper()
	user := os.Getenv("RTIFACT_REGISTRY_USER")
	pass := os.Getenv("RTIFACT_REGISTRY_PASSWORD")
	args := []string{
		"upgrade", "--install", release, e2eChartRef(t) + "/" + chart,
		"--version", version, "-n", ns, "--create-namespace", "--wait=hookOnly",
		"--kubeconfig", e2eKubeconfig(t), "--plain-http",
	}
	if user != "" {
		args = append(args, "--username", user, "--password", pass)
	}
	cmd := exec.Command("helm", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("helm upgrade --install %s: %v\n%s", release, err, out)
	}
}

func e2eAssertPlanStuckRecovery(t *testing.T, planOut, file, release string) {
	t.Helper()
	want := "rtifact deploy rollback"
	if !strings.Contains(planOut, want) {
		t.Fatalf("plan missing recovery (%q) in:\n%s", want, planOut)
	}
	if !strings.Contains(planOut, release) {
		t.Fatalf("plan missing release %q in:\n%s", release, planOut)
	}
	_ = filepath.Base(file)
}

func e2eRegistryHost(t *testing.T, file string) string {
	t.Helper()
	return e2eDoc(t, file).Registry.Host
}

func e2eAssertErrContains(t *testing.T, err error, subs ...string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	for _, s := range subs {
		if !strings.Contains(msg, s) {
			t.Fatalf("error %q missing %q", msg, s)
		}
	}
}

func e2eAssertIsInterrupted(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, deploy.ErrInterrupted) {
		t.Fatalf("expected ErrInterrupted, got %v", err)
	}
}

func e2eLockHeldInErr(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected lock error")
	}
	if !strings.Contains(err.Error(), "Lock held by") {
		t.Fatalf("expected lock holder in error: %v", err)
	}
}

func e2eIntEq(t *testing.T, label string, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %d want %d", label, got, want)
	}
}

func e2eDynamicClient(t *testing.T, kcfg kube.Config) dynamic.Interface {
	t.Helper()
	restCfg, err := kube.RESTConfig(kcfg)
	if err != nil {
		t.Fatal(err)
	}
	cs, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		t.Fatal(err)
	}
	return cs
}
