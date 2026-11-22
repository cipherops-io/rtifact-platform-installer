package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/paths"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
)

// layerDeploymentMarker uses saved stack outputs as a lightweight “still applied” signal.
func layerDeploymentMarker(prefix string, layer stacks.Layer) bool {
	path := filepath.Join(paths.ArtifactsDir(prefix), fmt.Sprintf("%s-outputs.json", layer))
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
	path := filepath.Join(paths.ArtifactsDir(prefix), fmt.Sprintf("%s-outputs.json", layer))
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func guardDestroyDependencies(opts Options, prefix string, layer stacks.Layer, batch []stacks.Layer) error {
	isDeployed := func(l stacks.Layer) bool {
		if opts.layerDeployed != nil {
			return opts.layerDeployed(l)
		}
		return layerDeploymentMarker(prefix, l)
	}
	down := stacks.Downstream(layer)
	var deps []string
	for _, d := range down {
		if downstreamAlreadyDestroyedInBatch(batch, layer, d) {
			continue
		}
		if isDeployed(d) {
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
	var cmds []string
	for _, d := range ordered {
		cmds = append(cmds, fmt.Sprintf("rtifact destroy <config.yaml> --stack %s", d))
	}
	cmds = append(cmds, fmt.Sprintf("rtifact destroy <config.yaml> --stack %s", layer))
	hint := strings.Join([]string{
		fmt.Sprintf("Cannot destroy stack %q: these stacks depend on it and are still deployed:", layer),
		"  • " + strings.Join(ordered, "\n  • "),
		"What you can do:",
		"  " + strings.Join(cmds, "\n  "),
		"or remove everything in order:",
		"  rtifact destroy <config.yaml> --stack all",
	}, "\n")
	return usererr.WithHint(
		fmt.Sprintf("Cannot destroy stack %q while dependent stacks are still deployed.", layer),
		hint,
		fmt.Sprintf("Dependent stacks still marked as deployed: %s", strings.Join(ordered, ", ")),
	)
}

func destroyDependencyGuard(opts Options, prefix string, layer stacks.Layer) error {
	if opts.Action != "destroy" {
		return nil
	}
	return guardDestroyDependencies(opts, prefix, layer, opts.DestroyBatch)
}