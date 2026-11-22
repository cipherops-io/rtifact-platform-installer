package golden

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud"
	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud/registry"
	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/paths"
	"github.com/cipherops-io/rtifact-platform-installer/internal/maputil"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
)

// Keys set by cloud CLI calls during a real run; omitted from layer tfvars goldens.
var excludedCloudKeys = []string{
	"resource_group_create",
	"resource_group_name",
}

type Snapshot struct {
	Merged     map[string]any
	Validation validationResult
	Backends   map[string]backendJSON
	Layers     map[string]layerJSON
}

type validationResult struct {
	Result string `json:"result"`
	Error  string `json:"error,omitempty"`
}

type backendJSON struct {
	Type     string            `json:"type"`
	Settings map[string]string `json:"settings"`
}

type layerJSON struct {
	// ExcludedCloudKeys documents vars omitted because they need live cloud checks (Azure RG).
	ExcludedCloudKeys []string       `json:"excluded_cloud_keys,omitempty"`
	Vars              map[string]any `json:"vars"`
}

func BuildSnapshot(repoRoot, configRel string) (*Snapshot, error) {
	configPath := filepath.Join(repoRoot, configRel)
	if err := config.ValidateCustomerYAML(configPath); err != nil {
		return &Snapshot{
			Validation: validationResult{Result: "fail", Error: usererr.Human(err)},
		}, nil
	}
	customer, err := config.LoadCustomer(configPath)
	if err != nil {
		return nil, err
	}
	profilesDir := filepath.Join(repoRoot, "profiles")
	defaults, err := config.LoadProfileDefaults(profilesDir, customer.Profile, customer.Cloud)
	if err != nil {
		return nil, err
	}
	meta, err := config.LoadProfileMeta(profilesDir, customer.Profile)
	if err != nil {
		return nil, err
	}
	merged, err := config.MergeBuildFromPath(defaults, configPath)
	if err != nil {
		return &Snapshot{
			Validation: validationResult{Result: "fail", Error: usererr.Human(err)},
		}, nil
	}
	provider, err := registry.Get(customer.Cloud)
	if err != nil {
		return nil, err
	}
	spec := config.CloudStateSpec(customer, meta.StackID)
	lifecycle, err := registry.LifecycleHooksFor(customer.Cloud)
	if err != nil {
		return nil, err
	}

	snap := &Snapshot{
		Merged:     merged,
		Validation: validationResult{Result: "pass"},
		Backends:   map[string]backendJSON{},
		Layers:     map[string]layerJSON{},
	}

	for _, layer := range stacks.AllLayers() {
		ls := layerSpec(spec, layer)
		backend, err := cloud.Backend(provider, ls, string(layer))
		if err != nil {
			return nil, fmt.Errorf("backend %s: %w", layer, err)
		}
		snap.Backends[string(layer)] = backendJSON{Type: backend.Type, Settings: backend.Settings}

		layerMerged := maputil.DeepCopyMap(merged)
		upstream := map[string]any{}
		for _, up := range stacks.Upstream(layer) {
			ub, err := cloud.Backend(provider, ls, string(up))
			if err != nil {
				return nil, fmt.Errorf("upstream backend %s: %w", up, err)
			}
			upstream[string(up)] = ub.RemoteState()
		}
		if len(upstream) > 0 {
			layerMerged["upstream_state"] = upstream
		}
		// injectAzureResourceGroupFlags calls Azure CLI — skipped for goldens; keys listed in ExcludedCloudKeys.
		edgeOut := filepath.Join(paths.ArtifactsDir(customer.NamePrefix()), string(stacks.LayerNetworkEdge)+"-outputs.json")
		lifecycle.MutateDNSDelegation(context.Background(), layer, customer.Account.Region, customer.NamePrefix(), edgeOut, layerMerged)
		for _, k := range excludedCloudKeys {
			delete(layerMerged, k)
		}
		lj := layerJSON{Vars: layerMerged}
		if customer.Cloud == "azure" && (layer == stacks.LayerStateBackend || layer == stacks.LayerCoreNetwork) {
			lj.ExcludedCloudKeys = append([]string(nil), excludedCloudKeys...)
		}
		snap.Layers[string(layer)] = lj
	}
	return snap, nil
}

func layerSpec(spec cloud.StateSpec, layer stacks.Layer) cloud.StateSpec {
	if layer == stacks.LayerStateBackend {
		spec.Local = true
		spec.LocalDir = filepath.Join(paths.ArtifactsDir(spec.NamePrefix), "state")
		_ = os.MkdirAll(spec.LocalDir, 0o755)
	}
	return spec
}

func (s *Snapshot) WriteDir(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	writes := []struct {
		name string
		val  any
	}{
		{"merged.json", s.Merged},
		{"validation.json", s.Validation},
		{"backends.json", s.Backends},
		{"layers.json", s.Layers},
	}
	for _, w := range writes {
		b, err := marshalSortedJSON(w.val)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, w.name), b, 0o644); err != nil {
			return err
		}
	}
	readme := `# Golden fixtures (W0)

Layer tfvars omit Azure resource_group_create and resource_group_name because
injectAzureResourceGroupFlags queries the live subscription (Azure CLI).
All other preparation matches internal/runner before WriteAutoTofuVarsJSON.
`
	return os.WriteFile(filepath.Join(dir, "README.md"), []byte(readme), 0o644)
}

func LoadSnapshot(dir string) (*Snapshot, error) {
	read := func(name string, dest any) error {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		return json.Unmarshal(b, dest)
	}
	var s Snapshot
	if err := read("merged.json", &s.Merged); err != nil {
		return nil, err
	}
	if err := read("validation.json", &s.Validation); err != nil {
		return nil, err
	}
	if err := read("backends.json", &s.Backends); err != nil {
		return nil, err
	}
	if err := read("layers.json", &s.Layers); err != nil {
		return nil, err
	}
	return &s, nil
}

func repoRoot(t testingTB) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}

type testingTB interface {
	Helper()
	Fatal(...any)
	Setenv(key, value string)
}

func initContentRoot(t testingTB) string {
	root := repoRoot(t)
	t.Setenv("INSTALLER_CONTENT_ROOT", root)
	return root
}