package tests

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func script(t *testing.T, name string) string {
	return filepath.Join(repoRoot(t), "build", "scripts", name)
}

func run(t *testing.T, env []string, name string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(script(t, name), args...)
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestMirrorProvidersVisitsEveryLayerForTargetPlatform(t *testing.T) {
	tmp := t.TempDir()
	log := filepath.Join(tmp, "tofu.log")
	fake := filepath.Join(tmp, "tofu")
	body := "#!/usr/bin/env bash\necho \"$(basename \"$PWD\") $*\" >> " + log + "\n"
	if err := os.WriteFile(fake, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, []string{
		"TOFU_BIN=" + fake, "GOOS=linux", "GOARCH=arm64",
		"PACKAGED_CLOUDS=aws", "MIRROR=" + filepath.Join(tmp, "mirror"),
	}, "mirror-providers.sh")
	if err != nil {
		t.Fatalf("mirror-providers.sh: %v\n%s", err, out)
	}
	raw, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	got := string(raw)

	layers, _ := filepath.Glob(filepath.Join(repoRoot(t), "infra", "stacks", "*", "aws", "*", "providers.tofu"))
	if len(layers) < 2 {
		t.Fatalf("expected several aws layers, found %d", len(layers))
	}
	for _, p := range layers {
		layer := filepath.Base(filepath.Dir(p))
		want := layer + " providers mirror -platform=linux_arm64 " + filepath.Join(tmp, "mirror")
		if !strings.Contains(got, want) {
			t.Errorf("layer %s not mirrored for linux_arm64; log:\n%s", layer, got)
		}
	}
	if n := strings.Count(got, "providers mirror"); n != len(layers) {
		t.Errorf("mirrored %d layers, want %d (only aws)", n, len(layers))
	}
}

func TestMirrorProvidersSkipsLayersAlreadyMirrored(t *testing.T) {
	tmp := t.TempDir()
	mirror := filepath.Join(tmp, "mirror")
	locks, _ := filepath.Glob(filepath.Join(repoRoot(t), "infra", "stacks", "*", "aws", "*", ".terraform.lock.hcl"))
	for _, l := range locks {
		raw, _ := os.ReadFile(l)
		var src string
		for _, line := range strings.Split(string(raw), "\n") {
			f := strings.Fields(line)
			switch {
			case len(f) >= 2 && f[0] == "provider":
				src = strings.Trim(f[1], `"`)
			case len(f) >= 3 && f[0] == "version" && src != "":
				ver := strings.Trim(f[2], `"`)
				name := src[strings.LastIndex(src, "/")+1:]
				writeFile(t, filepath.Join(mirror, src, "terraform-provider-"+name+"_"+ver+"_linux_amd64.zip"), "z", 0o644)
				src = ""
			}
		}
	}
	log := filepath.Join(tmp, "tofu.log")
	fake := filepath.Join(tmp, "tofu")
	writeFile(t, fake, "#!/usr/bin/env bash\necho \"$*\" >> "+log+"\n", 0o755)
	out, err := run(t, []string{"TOFU_BIN=" + fake, "GOOS=linux", "GOARCH=amd64", "PACKAGED_CLOUDS=aws", "MIRROR=" + mirror}, "mirror-providers.sh")
	if err != nil {
		t.Fatalf("mirror-providers.sh: %v\n%s", err, out)
	}
	if raw, _ := os.ReadFile(log); strings.Contains(string(raw), "providers mirror") {
		t.Errorf("mirrored again although every provider is present:\n%s", raw)
	}
	if !strings.Contains(out, "already mirrored") {
		t.Errorf("no skip message:\n%s", out)
	}
}

func TestMirrorProvidersRequiresHostTofu(t *testing.T) {
	out, err := run(t, []string{"TOFU_BIN=", "GOOS=linux", "GOARCH=amd64"}, "mirror-providers.sh")
	if err == nil || !strings.Contains(out, "TOFU_BIN") {
		t.Fatalf("want failure naming TOFU_BIN, got err=%v\n%s", err, out)
	}
}

func fakeTofuRelease(t *testing.T, dir, version, sum string) {
	t.Helper()
	rel := filepath.Join(dir, "v"+version)
	if err := os.MkdirAll(rel, 0o755); err != nil {
		t.Fatal(err)
	}
	name := "tofu_" + version + "_linux_amd64.zip"
	f, err := os.Create(filepath.Join(rel, name))
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, _ := zw.Create("tofu")
	_, _ = w.Write([]byte("#!/bin/sh\necho fake\n"))
	_ = zw.Close()
	_ = f.Close()
	data, _ := os.ReadFile(filepath.Join(rel, name))
	h := sha256.Sum256(data)
	if sum == "" {
		sum = hex.EncodeToString(h[:])
	}
	sums := sum + "  " + name + "\n"
	if err := os.WriteFile(filepath.Join(rel, "tofu_"+version+"_SHA256SUMS"), []byte(sums), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFetchTofuVerifiesChecksum(t *testing.T) {
	tmp := t.TempDir()
	good := filepath.Join(tmp, "good")
	fakeTofuRelease(t, good, "9.9.9", "")
	dest := filepath.Join(tmp, "dest-ok")
	out, err := run(t, []string{"TOFU_DOWNLOAD_BASE=file://" + good}, "fetch-tofu.sh", "9.9.9", "linux", "amd64", dest)
	if err != nil {
		t.Fatalf("fetch-tofu.sh with matching checksum: %v\n%s", err, out)
	}
	if st, err := os.Stat(filepath.Join(dest, "tofu")); err != nil || st.Mode()&0o111 == 0 {
		t.Fatalf("tofu not installed as executable: %v", err)
	}

	bad := filepath.Join(tmp, "bad")
	fakeTofuRelease(t, bad, "9.9.9", strings.Repeat("0", 64))
	dest = filepath.Join(tmp, "dest-bad")
	out, err = run(t, []string{"TOFU_DOWNLOAD_BASE=file://" + bad}, "fetch-tofu.sh", "9.9.9", "linux", "amd64", dest)
	if err == nil || !strings.Contains(out, "checksum mismatch") {
		t.Fatalf("want checksum mismatch, got err=%v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(dest, "tofu")); err == nil {
		t.Fatal("tofu installed despite checksum mismatch")
	}
}

func TestFetchTofuRejectsUnsupportedPlatformAndMissingArgs(t *testing.T) {
	if out, err := run(t, nil, "fetch-tofu.sh", "1.0.0"); err == nil {
		t.Fatalf("want usage error without platform and destination\n%s", out)
	}
	if out, err := run(t, nil, "fetch-tofu.sh", "1.0.0", "windows", "amd64", t.TempDir()); err == nil || !strings.Contains(out, "unsupported platform") {
		t.Fatalf("want unsupported platform, got err=%v\n%s", err, out)
	}
}

func writeFile(t *testing.T, path, body string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), mode); err != nil {
		t.Fatal(err)
	}
}

func TestCheckRelease(t *testing.T) {
	tmp := t.TempDir()
	embed := filepath.Join(tmp, "embed")
	stacks := filepath.Join(tmp, "stacks")
	bin := filepath.Join(tmp, "rtifact")
	writeFile(t, bin, "x", 0o755)
	writeFile(t, filepath.Join(stacks, "s", "aws", "core-network", ".terraform.lock.hcl"),
		"provider \"registry.opentofu.org/opentofu/aws\" {\n  version     = \"6.73.0\"\n}\n\nprovider \"registry.opentofu.org/hashicorp/tls\" {\n  version = \"4.4.2\"\n}\n", 0o644)
	env := []string{"CHECK_EMBED_DIR=" + embed, "CHECK_STACKS_DIR=" + stacks}
	check := func() (string, error) {
		return run(t, env, "check-release.sh", bin, "linux", "arm64", "aws", "0")
	}

	writeFile(t, filepath.Join(embed, "tofu", "linux_arm64", "tofu"), "x", 0o755)
	writeFile(t, filepath.Join(embed, "mirror", "registry.opentofu.org", "opentofu", "aws", "terraform-provider-aws_6.73.0_linux_arm64.zip"), "x", 0o644)
	out, err := check()
	if err == nil || !strings.Contains(out, "hashicorp/tls 4.4.2") {
		t.Fatalf("want missing tls provider, got err=%v\n%s", err, out)
	}

	writeFile(t, filepath.Join(embed, "mirror", "registry.opentofu.org", "hashicorp", "tls", "terraform-provider-tls_4.4.2_linux_arm64.zip"), "x", 0o644)
	if out, err := check(); err != nil {
		t.Fatalf("want OK, got %v\n%s", err, out)
	}

	writeFile(t, filepath.Join(embed, "tofu", "linux_amd64", "tofu"), "x", 0o755)
	if out, err := check(); err == nil || !strings.Contains(out, "another platform") {
		t.Fatalf("want failure for a second platform's OpenTofu, got err=%v\n%s", err, out)
	}
	_ = os.RemoveAll(filepath.Join(embed, "tofu", "linux_amd64"))

	if out, err := run(t, env, "check-release.sh", bin, "linux", "arm64", "aws", "1"); err == nil || !strings.Contains(out, "smaller than") {
		t.Fatalf("want size floor failure, got err=%v\n%s", err, out)
	}
}
