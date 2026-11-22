//go:build e2e

package deploy_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
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
		if err := deploy.Destroy(context.Background(), o); err != nil {
			t.Logf("cleanup destroy: %v", err)
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

func e2eInventory(t *testing.T, file string) []inventory.Entry {
	t.Helper()
	doc := e2eDoc(t, file)
	inv, err := inventory.Read(context.Background(), e2eK8sCfg(), deploy.DeploymentID(doc))
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

func e2eHelmRevision(t *testing.T, release, ns string) int {
	t.Helper()
	cmd := exec.Command("helm", "history", release, "-n", ns, "-o", "json",
		"--kubeconfig", e2eKubeconfig(t), "--max", "128")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("helm history %s/%s: %v", ns, release, err)
	}
	var rows []struct {
		Revision int `json:"revision"`
	}
	if err := json.Unmarshal(out, &rows); err != nil {
		t.Fatalf("helm history json: %v", err)
	}
	if len(rows) == 0 {
		t.Fatalf("no helm history for %s/%s", ns, release)
	}
	return rows[len(rows)-1].Revision
}

func e2eHelmReleaseMissing(t *testing.T, release, ns string) {
	t.Helper()
	cmd := exec.Command("helm", "status", release, "-n", ns, "--kubeconfig", e2eKubeconfig(t))
	if err := cmd.Run(); err == nil {
		t.Fatalf("expected helm release %s/%s to be absent", ns, release)
	}
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
		t.Fatalf("plan: code=%d err=%v", code, err)
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

func e2eChartRef(t *testing.T) string {
	t.Helper()
	host := os.Getenv("DEPLOY_E2E_REGISTRY_HOST")
	if host == "" {
		t.Fatal("DEPLOY_E2E_REGISTRY_HOST not set")
	}
	return "oci://" + host + "/charts"
}

func e2eHelmUpgradeNoWait(t *testing.T, release, ns, chart, version string) {
	t.Helper()
	user := os.Getenv("RTIFACT_REGISTRY_USER")
	pass := os.Getenv("RTIFACT_REGISTRY_PASSWORD")
	args := []string{
		"upgrade", "--install", release, e2eChartRef(t) + "/" + chart,
		"--version", version, "-n", ns, "--wait=false",
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