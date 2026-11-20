package stacks

// Layer is an installable OpenTofu stack under infra/stacks/<profile>/<cloud>/<layer>/.
type Layer string

const (
	LayerStateBackend  Layer = "state-backend"
	LayerCoreNetwork   Layer = "core-network"
	LayerDataPlatform  Layer = "data-platform"
	LayerAppPlatform   Layer = "app-platform"
	LayerNetworkEdge   Layer = "network-edge"
	LayerAIIntegration Layer = "ai-integrations"
	LayerAll           Layer = "all"
)

// ApplyOrder is the order of `--stack all`. state-backend is excluded: it is
// applied by `installer bootstrap` and must outlive every other layer.
var ApplyOrder = []Layer{
	LayerCoreNetwork,
	LayerDataPlatform,
	LayerAppPlatform,
	LayerNetworkEdge,
	LayerAIIntegration,
}

// Requires is the state-dependency graph: remote_state reads and plan upstream checks.
var Requires = map[Layer][]Layer{
	LayerCoreNetwork:   {},
	LayerDataPlatform:  {LayerCoreNetwork},
	LayerAppPlatform:   {LayerCoreNetwork},
	LayerNetworkEdge:   {LayerCoreNetwork},
	LayerAIIntegration: {LayerCoreNetwork, LayerDataPlatform},
}

// RuntimeRequires are ordering/safety edges that are not expressed as terraform_remote_state.
var RuntimeRequires = map[Layer][]Layer{
	LayerAppPlatform: {LayerDataPlatform},
	LayerNetworkEdge: {LayerAppPlatform},
}

// AllLayers returns every layer, state-backend first.
func AllLayers() []Layer {
	out := make([]Layer, 0, len(ApplyOrder)+1)
	out = append(out, LayerStateBackend)
	return append(out, ApplyOrder...)
}

// ParseLayer parses a --stack value.
func ParseLayer(s string) (Layer, bool) {
	l := Layer(s)
	if l == LayerAll {
		return l, true
	}
	for _, known := range AllLayers() {
		if l == known {
			return l, true
		}
	}
	return "", false
}

// ResolveLayers expands LayerAll to ApplyOrder; any other layer is returned as-is.
func ResolveLayers(stack Layer) []Layer {
	if stack == LayerAll {
		out := make([]Layer, len(ApplyOrder))
		copy(out, ApplyOrder)
		return out
	}
	return []Layer{stack}
}

// DestroyOrder is ApplyOrder reversed (state-backend is never part of `all`).
func DestroyOrder() []Layer {
	order := make([]Layer, len(ApplyOrder))
	for i, l := range ApplyOrder {
		order[len(ApplyOrder)-1-i] = l
	}
	return order
}

// DestroyOrderIndex returns the position of layer in DestroyOrder, or -1.
func DestroyOrderIndex(layer Layer) int {
	for i, l := range DestroyOrder() {
		if l == layer {
			return i
		}
	}
	return -1
}

// Upstream returns direct state dependencies for layer (terraform_remote_state targets).
func Upstream(layer Layer) []Layer {
	if layer == LayerStateBackend {
		return nil
	}
	deps := Requires[layer]
	if len(deps) == 0 {
		return nil
	}
	out := make([]Layer, len(deps))
	copy(out, deps)
	return out
}

// RuntimeUpstream returns runtime-only dependencies for apply/destroy policy.
func RuntimeUpstream(layer Layer) []Layer {
	deps := RuntimeRequires[layer]
	if len(deps) == 0 {
		return nil
	}
	out := make([]Layer, len(deps))
	copy(out, deps)
	return out
}

// UpstreamForPlan returns state + runtime dependencies when runtime checks are enabled.
func UpstreamForPlan(layer Layer, includeRuntime bool) []Layer {
	seen := map[Layer]bool{}
	var out []Layer
	add := func(ls []Layer) {
		for _, l := range ls {
			if seen[l] {
				continue
			}
			seen[l] = true
			out = append(out, l)
		}
	}
	add(Upstream(layer))
	if includeRuntime {
		add(RuntimeUpstream(layer))
	}
	return out
}

// Downstream returns layers that depend on layer (state and runtime edges) for destroy guards.
func Downstream(layer Layer) []Layer {
	if layer == LayerStateBackend {
		out := make([]Layer, len(ApplyOrder))
		copy(out, ApplyOrder)
		return out
	}
	var out []Layer
	for _, l := range ApplyOrder {
		if directDependsOn(l, layer, true) {
			out = append(out, l)
		}
	}
	return out
}

func directDependsOn(child, parent Layer, includeRuntime bool) bool {
	for _, r := range Requires[child] {
		if r == parent {
			return true
		}
	}
	if includeRuntime {
		for _, r := range RuntimeRequires[child] {
			if r == parent {
				return true
			}
		}
	}
	return false
}