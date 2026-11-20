package config_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type manifestFile struct {
	Path     string `yaml:"Path"`
	Checksum string `yaml:"Checksum"`
}

type boilerplateManifest struct {
	Files []manifestFile `yaml:"Files"`
}

// TestManagedManifestChecksums ensures on-disk managed files match boilerplate-manifest.yaml
// (replaces cross-repo verify-managed in CI without network access to boilerplates).
func TestManagedManifestChecksums(t *testing.T) {
	root := filepath.Join("..", "..")
	manifestPath := filepath.Join(root, "boilerplate-manifest.yaml")
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var m boilerplateManifest
	if err := yaml.Unmarshal(raw, &m); err != nil {
		t.Fatalf("parse manifest: %v", err)
	}
	if len(m.Files) == 0 {
		t.Fatal("manifest Files: empty (expected managed file checksums)")
	}
	for _, ent := range m.Files {
		if ent.Path == "" || ent.Checksum == "" {
			t.Fatalf("invalid manifest entry: %+v", ent)
		}
		want, ok := parseSHA256(ent.Checksum)
		if !ok {
			t.Fatalf("%s: bad checksum %q", ent.Path, ent.Checksum)
		}
		path := filepath.Join(root, ent.Path)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("%s: %v", ent.Path, err)
			continue
		}
		sum := sha256.Sum256(data)
		got := hex.EncodeToString(sum[:])
		if got != want {
			t.Errorf("%s: checksum mismatch\n  got:  sha256:%s\n  want: sha256:%s", ent.Path, got, want)
		}
	}
}

func parseSHA256(s string) (string, bool) {
	s = strings.TrimSpace(s)
	const p = "sha256:"
	if !strings.HasPrefix(s, p) {
		return "", false
	}
	h := strings.TrimPrefix(s, p)
	if len(h) != 64 {
		return "", false
	}
	for _, c := range h {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return "", false
		}
	}
	return h, true
}