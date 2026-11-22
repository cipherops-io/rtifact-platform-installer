package deploy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyDefaults_crds(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rel.yaml")
	if err := os.WriteFile(path, []byte(`schema_version: 1
kind: deploy
customer: { name: X, code: x }
environment: dev
name_prefix: x-dev
registry: { host: h, charts_path: p }
namespaces: [{ name: x }]
releases:
  - name: app
    namespace: x
    chart: c
    version: 1.0.0
`), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Releases[0].CRDs != "chart" {
		t.Fatalf("crds default: got %q", doc.Releases[0].CRDs)
	}
}

func TestApplyDefaults_noDefaultsBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rel.yaml")
	if err := os.WriteFile(path, []byte(`schema_version: 1
kind: deploy
customer: { name: X, code: x }
environment: dev
name_prefix: x-dev
registry: { host: h, charts_path: p }
namespaces: [{ name: x }]
releases:
  - name: app
    namespace: x
    chart: c
    version: 1.0.0
`), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Defaults.Timeout != "10m" {
		t.Fatalf("timeout: %q", doc.Defaults.Timeout)
	}
	if doc.Defaults.Wait == nil || !*doc.Defaults.Wait {
		t.Fatal("wait should default true")
	}
	if doc.Defaults.RollbackOnFailure == nil || !*doc.Defaults.RollbackOnFailure {
		t.Fatal("rollback_on_failure should default true")
	}
	if doc.Defaults.HistoryMax != 5 {
		t.Fatalf("history_max: %d", doc.Defaults.HistoryMax)
	}
}

func TestApplyDefaults_partialDefaultsBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rel.yaml")
	if err := os.WriteFile(path, []byte(`schema_version: 1
kind: deploy
customer: { name: X, code: x }
environment: dev
name_prefix: x-dev
registry: { host: h, charts_path: p }
defaults:
  wait: false
namespaces: [{ name: x }]
releases:
  - name: app
    namespace: x
    chart: c
    version: 1.0.0
`), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Defaults.Wait == nil || *doc.Defaults.Wait {
		t.Fatal("explicit wait: false should be kept")
	}
	if doc.Defaults.RollbackOnFailure == nil || !*doc.Defaults.RollbackOnFailure {
		t.Fatal("unset rollback_on_failure should default true")
	}
	if doc.Defaults.Timeout != "10m" {
		t.Fatalf("timeout default: %q", doc.Defaults.Timeout)
	}
}