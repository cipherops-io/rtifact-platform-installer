package runner

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud"
	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud/azure"
	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

func TestUpstreamReadyForPlan_NoUpstreamLayers(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	provider := azure.New()
	spec := cloud.StateSpec{NamePrefix: "test-noup"}
	_, ready, err := upstreamReadyForPlan(context.Background(),
		&config.CustomerConfig{Cloud: azure.CloudName},
		&config.ProfileMeta{StackID: "aios-infra"},
		provider, spec, repoRoot(t), map[string]any{}, stacks.LayerStateBackend,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !ready {
		t.Fatal("state-backend should plan without upstream layers")
	}
}

func TestUpstreamReadyForPlan_LayerSpecUsesLocalStateForStateBackend(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	spec := cloud.StateSpec{NamePrefix: "test-local-sb"}
	after := layerSpec(spec, stacks.LayerStateBackend)
	if !after.Local {
		t.Fatal("state-backend spec should use local backend")
	}
	provider := azure.New()
	ub, err := cloud.Backend(provider, layerSpec(spec, stacks.LayerStateBackend), string(stacks.LayerStateBackend))
	if err != nil {
		t.Fatal(err)
	}
	if ub.Type != "local" {
		t.Fatalf("backend type=%q", ub.Type)
	}
	want := filepath.Join(after.LocalDir, "state-backend.tfstate")
	if ub.Settings["path"] != want {
		t.Fatalf("local state path=%q want %q", ub.Settings["path"], want)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	return filepath.Join(filepath.Dir(file), "..", "..")
}