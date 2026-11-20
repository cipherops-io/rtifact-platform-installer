package config_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

var remoteStateLine = regexp.MustCompile(`data\s+"terraform_remote_state"\s+"([a-z0-9_]+)"`)

// remoteStateReads maps layer → expected upstream layer names embedded in remote_state data source suffixes.
// Contract: stack main*.tofu must match stacks.Requires (plus naming convention core_network → core-network).
var remoteStateReads = map[stacks.Layer][]stacks.Layer{
	stacks.LayerDataPlatform:  {stacks.LayerCoreNetwork},
	stacks.LayerAppPlatform:   {stacks.LayerCoreNetwork},
	stacks.LayerNetworkEdge:   {stacks.LayerCoreNetwork},
	stacks.LayerAIIntegration: {stacks.LayerCoreNetwork, stacks.LayerDataPlatform},
}

func remoteStateNames(layer stacks.Layer, cloud string) (map[string]bool, error) {
	dir := filepath.Join("../../infra/stacks/aios-infra", cloud, string(layer))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	found := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "main") || !strings.HasSuffix(e.Name(), ".tofu") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		for _, line := range strings.Split(string(b), "\n") {
			if m := remoteStateLine.FindStringSubmatch(strings.TrimSpace(line)); len(m) == 2 {
				found[m[1]] = true
			}
		}
	}
	return found, nil
}

func layerRemoteID(l stacks.Layer) string {
	return strings.ReplaceAll(string(l), "-", "_")
}

func TestRemoteStateGraphMatchesRequires(t *testing.T) {
	clouds := []string{"azure", "aws", "gcp"}
	for layer, want := range remoteStateReads {
		for _, cloud := range clouds {
			if layer == stacks.LayerAIIntegration && cloud != "azure" {
				// AWS/GCP ai-integrations stacks may only read core-network today.
				want = []stacks.Layer{stacks.LayerCoreNetwork}
			}
			got, err := remoteStateNames(layer, cloud)
			if err != nil {
				t.Fatalf("%s/%s: %v", cloud, layer, err)
			}
			wantNames := map[string]bool{}
			for _, up := range want {
				wantNames[layerRemoteID(up)] = true
			}
			if len(got) != len(wantNames) {
				t.Errorf("%s %s: remote_state reads %v want %v", cloud, layer, keys(got), keys(wantNames))
				continue
			}
			for n := range wantNames {
				if !got[n] {
					t.Errorf("%s %s: missing remote_state %q", cloud, layer, n)
				}
			}
			for n := range got {
				if !wantNames[n] {
					t.Errorf("%s %s: unexpected remote_state %q", cloud, layer, n)
				}
			}
		}
	}
	for layer := range stacks.Requires {
		if layer == stacks.LayerCoreNetwork {
			continue
		}
		if _, ok := remoteStateReads[layer]; !ok {
			t.Errorf("remoteStateReads missing layer %q", layer)
		}
	}
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}