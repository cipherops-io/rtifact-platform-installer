package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
)

// layerDeploymentMarker uses saved stack outputs as a lightweight “still applied” signal.
func layerDeploymentMarker(prefix string, layer stacks.Layer) bool {
	path := filepath.Join(config.ArtifactsDir(prefix), fmt.Sprintf("%s-outputs.json", layer))
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return false
	}
	for _, v := range out {
		if v != nil && v != "" {
			return true
		}
	}
	return false
}

// downstreamAlreadyDestroyedInBatch is true when layer is destroyed via --stack all
// and downstream was processed earlier in DestroyOrder in the same run.
func downstreamAlreadyDestroyedInBatch(batch []stacks.Layer, layer, downstream stacks.Layer) bool {
	if len(batch) == 0 {
		return false
	}
	if !containsLayer(batch, downstream) {
		return false
	}
	layerIdx, downIdx := stacks.DestroyOrderIndex(layer), stacks.DestroyOrderIndex(downstream)
	return layerIdx >= 0 && downIdx >= 0 && downIdx < layerIdx
}

func clearLayerOutputs(prefix string, layer stacks.Layer) error {
	path := filepath.Join(config.ArtifactsDir(prefix), fmt.Sprintf("%s-outputs.json", layer))
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func guardDestroyDependencies(prefix string, layer stacks.Layer, batch []stacks.Layer) error {
	down := stacks.Downstream(layer)
	var deps []string
	for _, d := range down {
		if downstreamAlreadyDestroyedInBatch(batch, layer, d) {
			continue
		}
		if layerDeploymentMarker(prefix, d) {
			deps = append(deps, string(d))
		}
	}
	if len(deps) == 0 {
		return nil
	}
	order := stacks.DestroyOrder()
	ordered := make([]string, 0, len(deps))
	for _, l := range order {
		for _, d := range deps {
			if string(l) == d {
				ordered = append(ordered, d)
				break
			}
		}
	}
	hint := fmt.Sprintf(
		"Remove dependent stacks first (recommended order: %s), then destroy %q.",
		strings.Join(ordered, " → "),
		layer,
	)
	return usererr.WithHint(
		fmt.Sprintf("Cannot destroy stack %q while dependent stacks are still deployed.", layer),
		hint,
		fmt.Sprintf("Dependent stacks still marked as applied: %s", strings.Join(ordered, ", ")),
	)
}

func destroyDependencyGuard(opts Options, prefix string, layer stacks.Layer) error {
	if opts.Action != "destroy" {
		return nil
	}
	return guardDestroyDependencies(prefix, layer, opts.DestroyBatch)
}