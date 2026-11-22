//go:build e2e

package deploy_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy"
)

func e2eKubeconfig(t *testing.T) string {
	t.Helper()
	k := os.Getenv("DEPLOY_E2E_KUBECONFIG")
	if k == "" {
		t.Fatal("DEPLOY_E2E_KUBECONFIG not set")
	}
	return k
}

func e2eFixture(t *testing.T) string {
	t.Helper()
	root := repoRoot(t)
	return filepath.Join(root, "tests", "deploy", "fixtures", "e2e.yaml")
}

func e2eOpts(t *testing.T, file string) deploy.RunOptions {
	t.Helper()
	return deploy.RunOptions{
		FilePath:      file,
		KubeconfigCLI: e2eKubeconfig(t),
		Yes:           true,
	}
}

func writeFixture(t *testing.T, dir string, mutate func(*deploy.Document)) string {
	t.Helper()
	path := e2eFixture(t)
	doc, err := deploy.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if exp := os.Getenv("DEPLOY_E2E_EXPECTED_SERVER"); exp != "" {
		doc.Cluster.ExpectedServer = exp
	}
	if host := os.Getenv("DEPLOY_E2E_REGISTRY_HOST"); host != "" {
		doc.Registry.Host = host
	}
	mutate(doc)
	isolateFixture(t, doc)
	out := filepath.Join(dir, "releases.yaml")
	if err := deploy.WriteDocument(out, doc); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestE2E_S1_FreshApply(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(*deploy.Document) {})
	o := e2eOpts(t, file)
	e2eCleanupDeploy(t, file, o)
	e2eMustApply(t, o)
	e2eAssertInventory(t, file, "ok-a", "ok-b")
	nsA := e2eReleaseNS(t, file, "ok-a")
	nsB := e2eReleaseNS(t, file, "ok-b")
	e2eIntEq(t, "ok-a revision", e2eHelmRevision(t, "ok-a", nsA), 1)
	e2eIntEq(t, "ok-b revision", e2eHelmRevision(t, "ok-b", nsB), 1)
}

func TestE2E_S2_ReapplyNoChanges(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(*deploy.Document) {})
	o := e2eOpts(t, file)
	e2eCleanupDeploy(t, file, o)
	e2eMustApply(t, o)
	nsA := e2eReleaseNS(t, file, "ok-a")
	revA := e2eHelmRevision(t, "ok-a", nsA)
	e2eMustApply(t, o)
	if got := e2eHelmRevision(t, "ok-a", nsA); got != revA {
		t.Fatalf("ok-a revision changed %d -> %d", revA, got)
	}
	o.DetailedExitCode = true
	code, out := e2ePlan(t, o)
	e2eIntEq(t, "plan exit", code, 0)
	if strings.Contains(out, "upgrade") {
		t.Fatalf("expected no changes in plan, got:\n%s", out)
	}
}

func TestE2E_S3_VersionBump(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(*deploy.Document) {})
	o := e2eOpts(t, file)
	e2eCleanupDeploy(t, file, o)
	e2eMustApply(t, o)
	nsA := e2eReleaseNS(t, file, "ok-a")
	rev0 := e2eHelmRevision(t, "ok-a", nsA)
	file = writeFixture(t, dir, func(doc *deploy.Document) {
		for i := range doc.Releases {
			if doc.Releases[i].Name == "ok-a" {
				doc.Releases[i].Version = "1.0.1"
			}
		}
	})
	o.FilePath = file
	o.DetailedExitCode = true
	code, out := e2ePlan(t, o)
	e2eIntEq(t, "plan exit", code, 2)
	if !strings.Contains(out, "upgrade") {
		t.Fatalf("expected upgrade in plan:\n%s", out)
	}
	e2eMustApply(t, o)
	e2eIntEq(t, "ok-a revision after bump", e2eHelmRevision(t, "ok-a", nsA), rev0+1)
}

func TestE2E_S4_ValueChange(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(*deploy.Document) {})
	o := e2eOpts(t, file)
	e2eCleanupDeploy(t, file, o)
	e2eMustApply(t, o)
	nsA := e2eReleaseNS(t, file, "ok-a")
	nsB := e2eReleaseNS(t, file, "ok-b")
	revB := e2eHelmRevision(t, "ok-b", nsB)
	revA0 := e2eHelmRevision(t, "ok-a", nsA)
	file = writeFixture(t, dir, func(doc *deploy.Document) {
		for i := range doc.Releases {
			if doc.Releases[i].Name == "ok-a" {
				doc.Releases[i].Values = map[string]any{"msg": "changed"}
			}
		}
	})
	o.FilePath = file
	o.DetailedExitCode = true
	code, out := e2ePlan(t, o)
	e2eIntEq(t, "plan exit", code, 2)
	if !strings.Contains(out, "ok-a") {
		t.Fatalf("expected ok-a in plan:\n%s", out)
	}
	if strings.Contains(out, "ok-b") && strings.Contains(out, "upgrade") {
		// ok-b should not need upgrade when only ok-a values change.
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, "ok-b") && strings.Contains(line, "upgrade") {
				t.Fatalf("unexpected ok-b upgrade:\n%s", out)
			}
		}
	}
	e2eMustApply(t, o)
	e2eIntEq(t, "ok-a revision", e2eHelmRevision(t, "ok-a", nsA), revA0+1)
	e2eIntEq(t, "ok-b revision unchanged", e2eHelmRevision(t, "ok-b", nsB), revB)
}

func TestE2E_S13_IdentityGuard(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Cluster.ExpectedServer = "https://wrong.example"
	})
	_, err := deploy.Plan(context.Background(), e2eOpts(t, file))
	e2eAssertErrContains(t, err, "wrong.example")
	e2eAssertInventory(t, file) // empty
}

func TestE2E_Canary_SecretNotInPlanOutput(t *testing.T) {
	const secret = "canary-e2e-secret-must-not-leak"
	t.Setenv("RTIFACT_E2E_SECRET", secret)
	dir := t.TempDir()
	file := writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Namespaces = append(doc.Namespaces, deploy.Namespace{Name: "sec-ns"})
		doc.Releases = append(doc.Releases, deploy.Release{
			Name: "secretful", Namespace: "sec-ns",
			Chart: "secretful", Version: "1.0.0", CRDs: "chart",
			Values: map[string]any{"token": "${env.RTIFACT_E2E_SECRET}"},
		})
	})
	var out, errOut bytes.Buffer
	o := e2eOpts(t, file)
	o.Stdout = &out
	o.Stderr = &errOut
	e2eCleanupDeploy(t, file, o)
	e2eMustApply(t, o)
	out.Reset()
	errOut.Reset()
	if _, err := deploy.Plan(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	combined := out.String() + errOut.String()
	if strings.Contains(combined, secret) {
		t.Fatalf("plan output leaked secret (len=%d)", len(combined))
	}
	if !strings.Contains(errOut.String(), "Identity:") {
		t.Fatal("expected Identity banner line on stderr")
	}
}

func TestE2E_S19_UnknownOnly(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(*deploy.Document) {})
	o := e2eOpts(t, file)
	o.Only = "does-not-exist"
	_, err := deploy.Plan(context.Background(), o)
	e2eAssertErrContains(t, err, "does-not-exist")
}

func TestE2E_S17_BadRegistryPassword(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("RTIFACT_REGISTRY_PASSWORD", "bad")
	file := writeFixture(t, dir, func(*deploy.Document) {})
	host := e2eRegistryHost(t, file)
	_, err := deploy.Plan(context.Background(), e2eOpts(t, file))
	e2eAssertErrContains(t, err, host)
	nsA := e2eReleaseNS(t, file, "ok-a")
	e2eHelmReleaseMissing(t, "ok-a", nsA)
}