package stacks

// Layer is an installable OpenTofu stack under infra/stacks/<profile>/<layer>/.
type Layer string

const (
	LayerBootstrap Layer = "bootstrap"
	LayerNetwork   Layer = "network"
	LayerData      Layer = "data"
	LayerCluster   Layer = "cluster"
	LayerAll       Layer = "all"
)

var ApplyOrder = []Layer{LayerNetwork, LayerData, LayerCluster}

func ParseLayer(s string) (Layer, bool) {
	switch Layer(s) {
	case LayerBootstrap, LayerNetwork, LayerData, LayerCluster, LayerAll:
		return Layer(s), true
	default:
		return "", false
	}
}

func ResolveLayers(stack Layer) []Layer {
	if stack == LayerAll {
		out := make([]Layer, len(ApplyOrder))
		copy(out, ApplyOrder)
		return out
	}
	return []Layer{stack}
}

func DestroyOrder() []Layer {
	order := make([]Layer, len(ApplyOrder))
	for i, l := range ApplyOrder {
		order[len(ApplyOrder)-1-i] = l
	}
	return order
}