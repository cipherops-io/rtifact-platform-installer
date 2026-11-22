package runner

import (
	"context"
	"sync"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

type fakeHookCloud struct {
	cloud.NoopLifecycle
	name string

	mu              sync.Mutex
	prepareCalls    int
	delegationCalls int
}

func (f *fakeHookCloud) Name() string { return f.name }

func (f *fakeHookCloud) EnsureStateBackend(context.Context, cloud.StateSpec) (*cloud.StateBackendInfo, error) {
	return nil, cloud.ErrNotImplemented
}

func (f *fakeHookCloud) Backend(cloud.StateSpec, string) (*cloud.BackendConfig, error) {
	return &cloud.BackendConfig{Type: "local", Settings: map[string]string{"path": "state.tfstate"}}, nil
}

func (f *fakeHookCloud) DefaultMaturity() string                  { return "operational" }
func (f *fakeHookCloud) CustomerSchemaUsesCommonAllOf() bool      { return true }
func (f *fakeHookCloud) DefaultQuotaCheckOnValidate() bool        { return false }
func (f *fakeHookCloud) TFVarSections() []string                  { return nil }
func (f *fakeHookCloud) ExpandCustomerPaths(string, map[string]any) error { return nil }
func (f *fakeHookCloud) Normalize(map[string]any)                 {}
func (f *fakeHookCloud) Validate(map[string]any) error            { return nil }
func (f *fakeHookCloud) Preflight(context.Context, map[string]any) error { return nil }

func (f *fakeHookCloud) PrepareLayerVars(ctx context.Context, _ cloud.StateSpec, _ stacks.Layer, _ map[string]any) error {
	f.mu.Lock()
	f.prepareCalls++
	f.mu.Unlock()
	return nil
}

func (f *fakeHookCloud) MutateDNSDelegation(ctx context.Context, _ stacks.Layer, _, _, _ string, _ map[string]any) {
	f.mu.Lock()
	f.delegationCalls++
	f.mu.Unlock()
}

type fakeTofuRunner struct {
	mu    sync.Mutex
	calls [][]string
}

func (f *fakeTofuRunner) Run(args ...string) error {
	f.mu.Lock()
	f.calls = append(f.calls, append([]string(nil), args...))
	f.mu.Unlock()
	return nil
}

func (f *fakeTofuRunner) Output(args ...string) ([]byte, error) {
	f.mu.Lock()
	f.calls = append(f.calls, append([]string(nil), args...))
	f.mu.Unlock()
	return []byte("{}"), nil
}

func (f *fakeTofuRunner) ShowJSON(string) ([]byte, error) {
	return []byte("{}"), nil
}

func TestHookWiringPrepareLayerMergedUsesCloudLifecycle(t *testing.T) {
	fc := &fakeHookCloud{name: "test"}
	merged := map[string]any{"features": map[string]any{}}
	spec := cloud.StateSpec{Profile: "aios-infra", NamePrefix: "demo"}
	_, err := prepareLayerMerged(context.Background(), fc, spec, stacks.LayerCoreNetwork, merged, upstreamSpecFlat)
	if err != nil {
		t.Fatal(err)
	}
	fc.mu.Lock()
	n := fc.prepareCalls
	fc.mu.Unlock()
	if n != 1 {
		t.Fatalf("PrepareLayerVars calls = %d, want 1", n)
	}
}

func TestHookWiringLayerPlanUsesTofuRunner(t *testing.T) {
	fake := &fakeTofuRunner{}
	lr := &layerRun{
		tofu:       fake,
		varFileArg: "-var-file=installer.auto.tfvars.json",
		opts:       Options{Action: "plan"},
	}
	if err := runLayerPlan(lr); err != nil {
		t.Fatal(err)
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if len(fake.calls) != 1 || fake.calls[0][0] != "plan" {
		t.Fatalf("tofu calls: %#v", fake.calls)
	}
}