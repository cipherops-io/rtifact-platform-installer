package deploy

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func repoRootFromTest(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestValidateFile_minimal(t *testing.T) {
	root := repoRootFromTest(t)
	t.Setenv("INSTALLER_CONTENT_ROOT", root)
	path := filepath.Join(root, "examples", "deploy", "minimal.yaml")
	if err := ValidateFile(path); err != nil {
		t.Fatal(err)
	}
}

func TestValidateFile_unknownKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
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
    typo_field: true
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateFile(path); err == nil {
		t.Fatal("expected schema rejection")
	}
}