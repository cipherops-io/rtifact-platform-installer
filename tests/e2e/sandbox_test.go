//go:build e2e

package e2e_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSandboxLifecycle(t *testing.T) {
	cfg := os.Getenv("E2E_CONFIG")
	if cfg == "" {
		t.Skip("E2E_CONFIG not set")
	}
	root, _ := os.Getwd()
	installer := filepath.Join(root, "..", "..", "dist", "installer")
	if _, err := os.Stat(installer); err != nil {
		t.Fatal("build installer first: make build-dev")
	}
	for _, args := range [][]string{
		{"validate", cfg, "--check-azure=false"},
		{"preflight", cfg},
		{"plan", cfg, "--stack", "network", "--yes"},
	} {
		cmd := exec.Command(installer, args...)
		cmd.Env = append(os.Environ(), "INSTALLER_CONTENT_ROOT="+filepath.Join(root, "..", ".."))
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%v: %s", args, out)
		}
	}
}