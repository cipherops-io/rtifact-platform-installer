package runner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

func TestGuardDestroyDependenciesBlocks(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	prefix := "test-guard"
	dir := config.ArtifactsDir(prefix)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	out := map[string]any{"aks_name": "demo-aks"}
	b, _ := json.Marshal(out)
	if err := os.WriteFile(filepath.Join(dir, "app-platform-outputs.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	err := guardDestroyDependencies(prefix, stacks.LayerDataPlatform, nil)
	if err == nil {
		t.Fatal("expected block when app-platform outputs exist")
	}
}

func TestGuardDestroyDependenciesAllows(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	prefix := "test-guard-ok"
	err := guardDestroyDependencies(prefix, stacks.LayerDataPlatform, nil)
	if err != nil {
		t.Fatalf("expected allow when no outputs: %v", err)
	}
}

func TestGuardDestroyDependenciesAllowsDownstreamAlreadyInDestroyBatch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	prefix := "test-guard-batch"
	dir := config.ArtifactsDir(prefix)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	out := map[string]any{"gateway_id": "demo-gw"}
	b, _ := json.Marshal(out)
	if err := os.WriteFile(filepath.Join(dir, "network-edge-outputs.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	batch := stacks.DestroyOrder()
	err := guardDestroyDependencies(prefix, stacks.LayerAppPlatform, batch)
	if err != nil {
		t.Fatalf("expected allow for app-platform when network-edge is in same destroy-all batch: %v", err)
	}
}