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

// Upstream returns the layers applied before layer in ApplyOrder (empty for
// state-backend and core-network).
func Upstream(layer Layer) []Layer {
	for i, l := range ApplyOrder {
		if l == layer {
			out := make([]Layer, i)
			copy(out, ApplyOrder[:i])
			return out
		}
	}
	return nil
}

// Downstream returns layers that are applied after layer in ApplyOrder and may
// read its remote state (destroy those before destroying layer).
func Downstream(layer Layer) []Layer {
	if layer == LayerStateBackend {
		out := make([]Layer, len(ApplyOrder))
		copy(out, ApplyOrder)
		return out
	}
	for i, l := range ApplyOrder {
		if l == layer {
			return append([]Layer(nil), ApplyOrder[i+1:]...)
		}
	}
	return nil
}
