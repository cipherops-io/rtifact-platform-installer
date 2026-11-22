//go:build e2e

package deploy_test

import (
	"context"
	"os"
	"path/filepath"
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
	out := filepath.Join(dir, "releases.yaml")
	if err := deploy.WriteDocument(out, doc); err != nil {
		t.Fatal(err)
	}
	return out
}

// S1 fresh apply ok-a, ok-b
func TestE2E_S1_FreshApply(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(*deploy.Document) {})
	if err := deploy.Apply(context.Background(), e2eOpts(t, file)); err != nil {
		t.Fatal(err)
	}
}

// S2 apply again — no changes
func TestE2E_S2_ReapplyNoChanges(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(*deploy.Document) {})
	o := e2eOpts(t, file)
	if err := deploy.Apply(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	code, err := deploy.Plan(context.Background(), o)
	if err != nil || code != 0 {
		t.Fatalf("plan after apply: code=%d err=%v", code, err)
	}
}

// S3 version bump shows upgrade
func TestE2E_S3_VersionBump(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(*deploy.Document) {})
	o := e2eOpts(t, file)
	_ = deploy.Apply(context.Background(), o)
	file = writeFixture(t, dir, func(doc *deploy.Document) {
		for i := range doc.Releases {
			if doc.Releases[i].Name == "ok-a" {
				doc.Releases[i].Version = "1.0.1"
			}
		}
	})
	o.FilePath = file
	o.DetailedExitCode = true
	code, err := deploy.Plan(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	if code != 2 {
		t.Fatalf("expected detailed-exitcode 2, got %d", code)
	}
}

// S4 value-only change upgrades ok-a only
func TestE2E_S4_ValueChange(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(*deploy.Document) {})
	o := e2eOpts(t, file)
	_ = deploy.Apply(context.Background(), o)
	file = writeFixture(t, dir, func(doc *deploy.Document) {
		for i := range doc.Releases {
			if doc.Releases[i].Name == "ok-a" {
				doc.Releases[i].Values = map[string]any{"msg": "changed"}
			}
		}
	})
	o.FilePath = file
	if err := deploy.Apply(context.Background(), o); err != nil {
		t.Fatal(err)
	}
}

// S13 wrong expected_server refused before cluster reads
func TestE2E_S13_IdentityGuard(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Cluster.ExpectedServer = "https://wrong.example"
	})
	_, err := deploy.Plan(context.Background(), e2eOpts(t, file))
	if err == nil {
		t.Fatal("expected identity refusal")
	}
}

// S17 wrong registry password
func TestE2E_S17_BadRegistryPassword(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("RTIFACT_REGISTRY_PASSWORD", "bad")
	file := writeFixture(t, dir, func(*deploy.Document) {})
	_, err := deploy.Plan(context.Background(), e2eOpts(t, file))
	if err == nil {
		t.Fatal("expected registry auth error")
	}
}