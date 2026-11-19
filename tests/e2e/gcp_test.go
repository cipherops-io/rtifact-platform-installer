//go:build e2e

package e2e_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGCPLifecycle(t *testing.T) {
	cfg := os.Getenv("E2E_GCP_CONFIG")
	if cfg == "" {
		t.Skip("E2E_GCP_CONFIG not set")
	}
	root, _ := os.Getwd()
	bin := filepath.Join(root, "..", "..", "dist", "rtifact")
	if _, err := os.Stat(bin); err != nil {
		t.Fatal("build rtifact first: make build-dev")
	}
	for _, args := range [][]string{
		{"gcp", "validate", cfg, "--check-quota=false"},
		{"gcp", "preflight", cfg},
	} {
		cmd := exec.Command(bin, args...)
		cmd.Env = append(os.Environ(), "INSTALLER_CONTENT_ROOT="+filepath.Join(root, "..", ".."))
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%v: %s", args, out)
		}
	}
}