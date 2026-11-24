package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fakeLock = `provider "registry.opentofu.org/opentofu/aws" {
  version     = "6.73.0"
  constraints = "6.73.0"
}

provider "registry.opentofu.org/hashicorp/tls" {
  version = "4.4.2"
}
`

func writeTestFile(t *testing.T, path, body string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), mode); err != nil {
		t.Fatal(err)
	}
}

// fakeDoctor builds a content root with one aws layer, a full mirror for linux_arm64 and a fake tofu.
func fakeDoctor(t *testing.T) (doctorEnv, *bytes.Buffer, string) {
	t.Helper()
	root := t.TempDir()
	layer := filepath.Join(root, "infra", "stacks", "s", "aws", "core-network")
	writeTestFile(t, filepath.Join(layer, "providers.tofu"), "terraform {}\n", 0o644)
	writeTestFile(t, filepath.Join(layer, ".terraform.lock.hcl"), fakeLock, 0o644)
	mirror := filepath.Join(root, "build", "embed", "mirror")
	writeTestFile(t, filepath.Join(mirror, "registry.opentofu.org/opentofu/aws/terraform-provider-aws_6.73.0_linux_arm64.zip"), "z", 0o644)
	writeTestFile(t, filepath.Join(mirror, "registry.opentofu.org/hashicorp/tls/terraform-provider-tls_4.4.2_linux_arm64.zip"), "z", 0o644)

	log := filepath.Join(root, "tofu.log")
	tofu := filepath.Join(root, "bin", "tofu")
	writeTestFile(t, tofu, `#!/bin/sh
if [ "$1" = version ]; then echo '{"terraform_version":"1.13.9"}'; exit 0; fi
echo "$PWD $* rc=$TF_CLI_CONFIG_FILE" >> `+log+`
[ -f "`+root+`/fail-init" ] && { echo "Error: provider not in mirror"; exit 1; }
exit 0
`, 0o755)

	out := &bytes.Buffer{}
	return doctorEnv{
		out:         out,
		mode:        "bundled",
		contentRoot: root,
		tofu:        tofu,
		mirror:      mirror,
		clouds:      []string{"aws"},
		wantTofu:    "1.13.9",
		platform:    "linux_arm64",
		lookPath:    func(string) (string, error) { return "", errors.New("not found") },
	}, out, log
}

func TestDoctorDeepPasses(t *testing.T) {
	d, out, log := fakeDoctor(t)
	if err := d.run(true); err != nil {
		t.Fatalf("doctor: %v\n%s", err, out)
	}
	for _, want := range []string{"ok   1.13.9", "ok   registry.opentofu.org/opentofu/aws 6.73.0", "ok   aws/core-network", "info gcloud not installed", "all checks passed"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
	raw, _ := os.ReadFile(log)
	if !strings.Contains(string(raw), "init -backend=false") || !strings.Contains(string(raw), "rc=") {
		t.Errorf("tofu init not run with a CLI config: %s", raw)
	}
	if strings.Contains(string(raw), d.contentRoot+"/infra") {
		t.Errorf("init ran inside the content root instead of a scratch copy: %s", raw)
	}
}

func TestDoctorReportsGaps(t *testing.T) {
	t.Run("provider missing from mirror", func(t *testing.T) {
		d, out, _ := fakeDoctor(t)
		_ = os.Remove(filepath.Join(d.mirror, "registry.opentofu.org/hashicorp/tls/terraform-provider-tls_4.4.2_linux_arm64.zip"))
		if err := d.run(false); err == nil || !strings.Contains(out.String(), "hashicorp/tls 4.4.2 missing from the mirror") {
			t.Fatalf("want missing provider, got %v\n%s", err, out)
		}
	})
	t.Run("mirror for another platform only", func(t *testing.T) {
		d, out, _ := fakeDoctor(t)
		d.platform = "darwin_amd64"
		if err := d.run(false); err == nil || !strings.Contains(out.String(), "missing from the mirror") {
			t.Fatalf("want missing providers for darwin_amd64, got %v\n%s", err, out)
		}
	})
	t.Run("wrong OpenTofu version", func(t *testing.T) {
		d, out, _ := fakeDoctor(t)
		d.wantTofu = "1.14.0"
		if err := d.run(false); err == nil || !strings.Contains(out.String(), "want 1.14.0") {
			t.Fatalf("want version mismatch, got %v\n%s", err, out)
		}
	})
	t.Run("offline init fails", func(t *testing.T) {
		d, out, _ := fakeDoctor(t)
		writeTestFile(t, filepath.Join(d.contentRoot, "fail-init"), "", 0o644)
		if err := d.run(true); err == nil || !strings.Contains(out.String(), "aws/core-network: tofu init failed") {
			t.Fatalf("want init failure, got %v\n%s", err, out)
		}
	})
	t.Run("bundled without mirror", func(t *testing.T) {
		d, out, _ := fakeDoctor(t)
		d.mirror = ""
		if err := d.run(false); err == nil || !strings.Contains(out.String(), "no provider mirror") {
			t.Fatalf("want missing mirror, got %v\n%s", err, out)
		}
	})
}

func TestDoctorDevMode(t *testing.T) {
	d, out, _ := fakeDoctor(t)
	d.mode, d.mirror, d.wantTofu = "dev", "", ""
	if err := d.run(false); err != nil {
		t.Fatalf("dev build without mirror should pass: %v\n%s", err, out)
	}
	out.Reset()
	if err := d.run(true); err == nil {
		t.Fatalf("--deep without a mirror should fail\n%s", out)
	}
}

func TestVersionJSON(t *testing.T) {
	var buf bytes.Buffer
	info := versionInfo{Version: "v0.0.1", Commit: "abc1234", Clouds: []string{"aws"}, Mode: "bundled", Platform: "linux/arm64", TofuVersion: "1.13.9", DeployEngine: "Helm v4.3.2"}
	if err := writeVersion(&buf, info, "json"); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"version", "commit", "clouds", "mode", "platform", "tofu_version"} {
		if _, ok := got[k]; !ok {
			t.Errorf("version -o json lacks %q: %s", k, buf.String())
		}
	}
	if _, ok := got["customer"]; ok {
		t.Error("version -o json still has customer")
	}
	if err := writeVersion(&buf, info, "yaml"); err == nil {
		t.Error("unknown format accepted")
	}
	buf.Reset()
	_ = writeVersion(&buf, info, "text")
	if !strings.Contains(buf.String(), "deploy-engine: Helm v4.3.2") || !strings.Contains(buf.String(), "mode=bundled") {
		t.Errorf("text output: %s", buf.String())
	}
}

func TestCurrentVersionDevBuild(t *testing.T) {
	info := currentVersion()
	if info.Mode != "dev" || info.Platform == "" || info.Clouds == nil {
		t.Errorf("unexpected dev version info: %+v", info)
	}
}
