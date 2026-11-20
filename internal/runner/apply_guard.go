package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
)

func previewCloudGate(cloud string, accept bool) error {
	if cloud == "azure" || accept {
		return nil
	}
	return usererr.WithHint(
		fmt.Sprintf("Cloud %q is preview in this installer build.", cloud),
		"Re-run with --accept-preview after reading docs; Azure is fully supported.",
		"Preview exit criteria: AWS/GCP stateful plan guard, corrected stack defaults, EnsureStateBackend without ErrNotImplemented, one full apply+destroy on a real account.",
	)
}

func applyUpstreamGuard(opts Options, layer stacks.Layer, deployed func(stacks.Layer) bool, configPath string) error {
	if opts.Action != "apply" || opts.Stack == stacks.LayerAll {
		return nil
	}
	missing := stacks.UpstreamForPlan(layer, !opts.IgnoreRuntimeDeps)
	for _, up := range missing {
		if deployed == nil || deployed(up) {
			continue
		}
		return usererr.WithHint(
			fmt.Sprintf("Cannot apply stack %q: required stack %q is not deployed yet.", layer, up),
			fmt.Sprintf("Run: rtifact apply %s --stack %s\nor: rtifact apply %s --stack all", configPath, up, configPath),
			"State dependencies must exist before this layer can apply.",
		)
	}
	return nil
}

type planOutputChange struct {
	ResourceChanges []struct {
		Type   string `json:"type"`
		Name   string `json:"name"`
		Change struct {
			Actions       []string `json:"actions"`
			Before        any      `json:"before"`
			After         any      `json:"after"`
			BeforeUnknown bool     `json:"before_unknown"`
			AfterUnknown  bool     `json:"after_unknown"`
		} `json:"change"`
	} `json:"resource_changes"`
	OutputChanges []struct {
		Address string `json:"address"`
		Actions []string `json:"actions"`
		Before  any      `json:"before"`
		After   any      `json:"after"`
	} `json:"output_changes"`
}

func outputChangeGuard(opts Options, layer stacks.Layer, planJSON []byte, deployed func(stacks.Layer) bool) error {
	if opts.Action != "apply" || opts.Stack == stacks.LayerAll || len(planJSON) == 0 {
		return nil
	}
	var p planOutputChange
	if err := json.Unmarshal(planJSON, &p); err != nil {
		return nil
	}
	if len(p.OutputChanges) == 0 {
		return nil
	}
	dependents := stacks.Downstream(layer)
	var risky []string
	for _, dep := range dependents {
		if deployed != nil && !deployed(dep) {
			continue
		}
		for _, oc := range p.OutputChanges {
			if outputChangeBreaking(oc.Actions) {
				risky = append(risky, fmt.Sprintf("output %q would change for stack %q (dependent %q may break)", strings.TrimPrefix(oc.Address, "output."), layer, dep))
			}
		}
	}
	if len(risky) == 0 {
		return nil
	}
	return usererr.WithHint(
		fmt.Sprintf("Apply to stack %q would change outputs read by deployed dependents.", layer),
		"Run apply with --stack all so dependent stacks stay consistent.",
		strings.Join(risky, "; "),
	)
}

func outputChangeBreaking(actions []string) bool {
	for _, a := range actions {
		switch a {
		case "delete", "update", "create":
			return true
		}
	}
	return false
}

func destroyDryRunReport(opts Options, prefix string, layers []stacks.Layer, deployed func(stacks.Layer) bool) error {
	if !opts.DryRunDestroy {
		return nil
	}
	fmt.Fprintf(os.Stderr, "Destroy dry-run for prefix %q (%d stacks):\n", prefix, len(layers))
	for _, layer := range layers {
		err := guardDestroyDependencies(opts, prefix, layer, layers)
		status := "OK"
		if err != nil {
			status = err.Error()
		}
		fmt.Fprintf(os.Stderr, "  • %s — guard: %s\n", layer, status)
	}
	fmt.Fprintln(os.Stderr, "destroy --dry-run: no changes applied")
	return nil
}