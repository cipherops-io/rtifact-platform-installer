package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func init() {
	if os.Getenv("INSTALLER_CONTENT_ROOT") != "" {
		return
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	_ = os.Setenv("INSTALLER_CONTENT_ROOT", root)
}

func minimalDeployYAML(extra string) string {
	base := `schema_version: 1
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
`
	return base + extra
}

func writeDeploy(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestValidate_removeRequiredDependency_S8(t *testing.T) {
	dir := t.TempDir()
	content := `schema_version: 1
kind: deploy
customer: { name: X, code: x }
environment: dev
name_prefix: x-dev
registry: { host: h, charts_path: p }
namespaces: [{ name: a }, { name: b }]
releases:
  - name: ok-b
    namespace: b
    chart: ok-b
    version: 1.0.0
    requires: [ok-a]
`
	p := writeDeploy(t, dir, "d.yaml", content)
	if err := ValidateFile(p); err == nil || !strings.Contains(err.Error(), "ok-a") {
		t.Fatalf("expected missing require ok-a, got %v", err)
	}
}

func TestValidate_unknownRequire(t *testing.T) {
	dir := t.TempDir()
	p := writeDeploy(t, dir, "d.yaml", minimalDeployYAML(`    requires: [missing]
`))
	if err := ValidateFile(p); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("expected unknown require error, got %v", err)
	}
}

func TestValidate_namespaceNotListed(t *testing.T) {
	dir := t.TempDir()
	yaml := strings.Replace(minimalDeployYAML(""), "namespace: x", "namespace: other", 1)
	p := writeDeploy(t, dir, "d.yaml", yaml)
	if err := ValidateFile(p); err == nil || !strings.Contains(err.Error(), "other") {
		t.Fatalf("expected namespace error, got %v", err)
	}
}

func TestValidate_missingValuesFile(t *testing.T) {
	dir := t.TempDir()
	p := writeDeploy(t, dir, "d.yaml", minimalDeployYAML(`    values_files: [no-such.yaml]
`))
	if err := ValidateFile(p); err == nil || !strings.Contains(err.Error(), "values file") {
		t.Fatalf("expected values file error, got %v", err)
	}
}

func TestValidate_duplicateReleaseName(t *testing.T) {
	dir := t.TempDir()
	content := `schema_version: 1
kind: deploy
customer: { name: X, code: x }
environment: dev
name_prefix: x-dev
registry: { host: h, charts_path: p }
namespaces: [{ name: x }]
releases:
  - name: dup
    namespace: x
    chart: a
    version: 1.0.0
  - name: dup
    namespace: x
    chart: b
    version: 1.0.0
`
	p := writeDeploy(t, dir, "d.yaml", content)
	if err := ValidateFile(p); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate release error, got %v", err)
	}
}

func TestValidate_defaultsTimeoutInvalid(t *testing.T) {
	dir := t.TempDir()
	yaml := strings.Replace(minimalDeployYAML(""), "name_prefix: x-dev", `name_prefix: x-dev
defaults:
  timeout: not-a-duration
`, 1)
	p := writeDeploy(t, dir, "d.yaml", yaml)
	if err := ValidateFile(p); err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

func TestValidate_releaseVersionNotSemver(t *testing.T) {
	dir := t.TempDir()
	yaml := strings.Replace(minimalDeployYAML(""), "version: 1.0.0", "version: latest", 1)
	p := writeDeploy(t, dir, "d.yaml", yaml)
	if err := ValidateFile(p); err == nil {
		t.Fatalf("expected semver error, got nil")
	}
}

func TestValidate_chartRefInvalidOCI(t *testing.T) {
	dir := t.TempDir()
	yaml := strings.Replace(minimalDeployYAML(""), "chart: c", "chart: oci://", 1)
	p := writeDeploy(t, dir, "d.yaml", yaml)
	if err := ValidateFile(p); err == nil || !strings.Contains(err.Error(), "oci://") {
		t.Fatalf("expected chart error, got %v", err)
	}
}

func TestValidate_pullSecretRequiresRegistryAuth(t *testing.T) {
	dir := t.TempDir()
	content := `schema_version: 1
kind: deploy
customer: { name: X, code: x }
environment: dev
name_prefix: x-dev
registry: { host: h, charts_path: p }
namespaces:
  - name: x
    pull_secret: true
releases:
  - name: app
    namespace: x
    chart: c
    version: 1.0.0
`
	p := writeDeploy(t, dir, "d.yaml", content)
	if err := ValidateFile(p); err == nil || !strings.Contains(err.Error(), "registry.auth") {
		t.Fatalf("expected registry.auth error, got %v", err)
	}
}

func TestValidate_attachPullSecretRequiresPullSecret(t *testing.T) {
	dir := t.TempDir()
	content := `schema_version: 1
kind: deploy
customer: { name: X, code: x }
environment: dev
name_prefix: x-dev
registry:
  host: h
  charts_path: p
  auth: { username_env: U, password_env: P }
namespaces:
  - name: x
    attach_pull_secret_to_default_sa: true
releases:
  - name: app
    namespace: x
    chart: c
    version: 1.0.0
`
	p := writeDeploy(t, dir, "d.yaml", content)
	if err := ValidateFile(p); err == nil || !strings.Contains(err.Error(), "attach_pull_secret") {
		t.Fatalf("expected attach error, got %v", err)
	}
}

func TestValidate_deploymentNameTooLong(t *testing.T) {
	dir := t.TempDir()
	longEnv := strings.Repeat("a", 62)
	content := fmt.Sprintf(`schema_version: 1
kind: deploy
customer: { name: X, code: x }
environment: %s
name_prefix: x-dev
registry: { host: h, charts_path: p }
namespaces: [{ name: x }]
releases:
  - name: app
    namespace: x
    chart: c
    version: 1.0.0
`, longEnv)
	p := writeDeploy(t, dir, "d.yaml", content)
	if err := ValidateFile(p); err == nil || !strings.Contains(err.Error(), "63") {
		t.Fatalf("expected length error, got %v", err)
	}
}

func TestValidate_duplicateNamespace(t *testing.T) {
	dir := t.TempDir()
	content := `schema_version: 1
kind: deploy
customer: { name: X, code: x }
environment: dev
name_prefix: x-dev
registry: { host: h, charts_path: p }
namespaces:
  - name: x
  - name: x
releases:
  - name: app
    namespace: x
    chart: c
    version: 1.0.0
`
	p := writeDeploy(t, dir, "d.yaml", content)
	err := ValidateFile(p)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "duplicate namespace") && !strings.Contains(err.Error(), "uniqueItems") {
		t.Fatalf("expected duplicate namespace error, got %v", err)
	}
}