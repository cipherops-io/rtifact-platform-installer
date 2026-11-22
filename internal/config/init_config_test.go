package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteInitConfigValidates(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "cfg.yaml")
	if err := WriteInitConfig("aws", "sandbox", "Acme", out); err != nil {
		t.Fatal(err)
	}
	if err := ValidateCustomerYAML(out); err != nil {
		t.Fatalf("init output should validate: %v", err)
	}
	raw, _ := os.ReadFile(out)
	if !containsAll(string(raw), "schema_version: 1", "cloud: aws", "tier: sandbox") {
		t.Fatalf("unexpected init output:\n%s", raw)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !stringsContains(s, p) {
			return false
		}
	}
	return true
}

func stringsContains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}