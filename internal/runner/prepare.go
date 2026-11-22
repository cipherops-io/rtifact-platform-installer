package runner

import (
	"context"
	"fmt"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud"
	"github.com/cipherops-io/rtifact-platform-installer/internal/maputil"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

// upstreamStateSpec picks the state spec used when resolving an upstream layer backend.
type upstreamStateSpec func(cloud.StateSpec, stacks.Layer) cloud.StateSpec

// prepareLayerMerged copies merged config, attaches upstream remote state, and
// runs cloud-specific layer var preparation.
func prepareLayerMerged(ctx context.Context, provider cloud.Provider, lifecycle cloud.Lifecycle, spec cloud.StateSpec, layer stacks.Layer, merged map[string]any, upSpec upstreamStateSpec) (map[string]any, error) {
	if upSpec == nil {
		upSpec = layerSpec
	}
	layerMerged := maputil.DeepCopyMap(merged)
	upstream := map[string]any{}
	for _, up := range stacks.Upstream(layer) {
		ub, err := cloud.Backend(provider, upSpec(spec, up), string(up))
		if err != nil {
			return nil, fmt.Errorf("upstream backend %s: %w", up, err)
		}
		upstream[string(up)] = ub.RemoteState()
	}
	if len(upstream) > 0 {
		layerMerged["upstream_state"] = upstream
	}
	if err := lifecycle.PrepareLayerVars(ctx, layerSpec(spec, layer), layer, layerMerged); err != nil {
		return nil, err
	}
	return layerMerged, nil
}

func upstreamSpecPerLayer(spec cloud.StateSpec, up stacks.Layer) cloud.StateSpec {
	return layerSpec(spec, up)
}

func upstreamSpecFlat(spec cloud.StateSpec, _ stacks.Layer) cloud.StateSpec {
	return spec
}