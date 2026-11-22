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