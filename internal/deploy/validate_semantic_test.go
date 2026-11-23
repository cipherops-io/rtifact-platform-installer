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
	wantErrContaining(t, ValidateFile(p), "ok-a")
}

func TestValidate_unknownRequire(t *testing.T) {
	dir := t.TempDir()
	p := writeDeploy(t, dir, "d.yaml", minimalDeployYAML(`    requires: [missing]
`))
	wantErrContaining(t, ValidateFile(p), "missing")
}

func TestValidate_namespaceNotListed(t *testing.T) {
	dir := t.TempDir()
	yaml := strings.Replace(minimalDeployYAML(""), "namespace: x", "namespace: other", 1)
	p := writeDeploy(t, dir, "d.yaml", yaml)
	wantErrContaining(t, ValidateFile(p), "other")
}

func TestValidate_missingValuesFile(t *testing.T) {
	dir := t.TempDir()
	p := writeDeploy(t, dir, "d.yaml", minimalDeployYAML(`    values_files: [no-such.yaml]
`))
	wantErrContaining(t, ValidateFile(p), "values file")
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
	wantErrContaining(t, ValidateFile(p), "duplicate")
}

func TestValidate_defaultsTimeoutInvalid(t *testing.T) {
	dir := t.TempDir()
	yaml := strings.Replace(minimalDeployYAML(""), "name_prefix: x-dev", `name_prefix: x-dev
defaults:
  timeout: not-a-duration
`, 1)
	p := writeDeploy(t, dir, "d.yaml", yaml)
	wantErrContaining(t, ValidateFile(p), "timeout")
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
	wantErrContaining(t, ValidateFile(p), "oci://")
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
	wantErrContaining(t, ValidateFile(p), "registry.auth")
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
	wantErrContaining(t, ValidateFile(p), "attach_pull_secret")
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
	wantErrContaining(t, ValidateFile(p), "63")
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
	wantErrContaining(t, err, "duplicate namespace", "uniqueItems")
}
