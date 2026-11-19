package config_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud/azure"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

const stackRoot = "../../infra/stacks/aios-infra/azure"

func TestEveryLayerHasStack(t *testing.T) {
	for _, l := range stacks.AllLayers() {
		dir := filepath.Join(stackRoot, string(l))
		for _, f := range []string{"main.tofu", "providers.tofu", "variables.tofu", ".terraform.lock.hcl"} {
			if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
				t.Errorf("layer %s: %v", l, err)
			}
		}
	}
	entries, _ := os.ReadDir(stackRoot)
	if len(entries) != len(stacks.AllLayers()) {
		t.Errorf("stack dirs = %d, layers = %d", len(entries), len(stacks.AllLayers()))
	}
}

func TestProviderPinsMatchVersions(t *testing.T) {
	for _, l := range stacks.AllLayers() {
		b, err := os.ReadFile(filepath.Join(stackRoot, string(l), "providers.tofu"))
		if err != nil {
			t.Fatal(err)
		}
		s := string(b)
		for _, want := range []string{
			`required_version = ">=`,
			`version = "` + azure.AzureRMProviderVersion + `"`,
			`version = "` + azure.RandomProviderVersion + `"`,
			`version = "` + azure.TLSProviderVersion + `"`,
		} {
			if !strings.Contains(s, want) {
				t.Errorf("layer %s providers.tofu missing %s", l, want)
			}
		}
	}
}

func TestNoLegacyInputFallbacks(t *testing.T) {
	_ = filepath.WalkDir(stackRoot, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".tofu") {
			return err
		}
		b, _ := os.ReadFile(p)
		for _, banned := range []string{"layer_inputs", "use_embedded_network"} {
			if strings.Contains(string(b), banned) {
				t.Errorf("%s still references %s", p, banned)
			}
		}
		return nil
	})
}

func TestOnlyCoreNetworkCreatesNaming(t *testing.T) {
	for _, l := range stacks.AllLayers() {
		hasNaming := false
		_ = filepath.WalkDir(filepath.Join(stackRoot, string(l)), func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(p, ".tofu") {
				return err
			}
			b, _ := os.ReadFile(p)
			if strings.Contains(string(b), `module "naming"`) {
				hasNaming = true
			}
			return nil
		})
		if hasNaming != (l == stacks.LayerCoreNetwork) {
			t.Errorf("layer %s: module naming present = %v", l, hasNaming)
		}
	}
}

func TestDownstreamLayersUseRemoteState(t *testing.T) {
	for _, l := range stacks.ApplyOrder[1:] {
		b, err := os.ReadFile(filepath.Join(stackRoot, string(l), "main.tofu"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(b), `data "terraform_remote_state" "core_network"`) {
			t.Errorf("layer %s does not read core-network remote state", l)
		}
	}
}

var moduleSource = regexp.MustCompile(`source\s*=\s*"(\.[^"]+)"`)

func TestModuleSourcesResolve(t *testing.T) {
	_ = filepath.WalkDir(stackRoot, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".tofu") {
			return err
		}
		b, _ := os.ReadFile(p)
		for _, m := range moduleSource.FindAllStringSubmatch(string(b), -1) {
			target := filepath.Join(filepath.Dir(p), m[1])
			if st, err := os.Stat(target); err != nil || !st.IsDir() {
				t.Errorf("%s: module source %s does not resolve", p, m[1])
			}
		}
		return nil
	})
}

func TestStateBackendModuleNoStringPlus(t *testing.T) {
	b, err := os.ReadFile("../../infra/modules/azure/state-backend/main.tofu")
	if err != nil {
		t.Fatal(err)
	}
	if regexp.MustCompile(`\)\s*\+\s*"`).Match(b) {
		t.Fatal("state-backend main.tofu concatenates strings with +")
	}
}
