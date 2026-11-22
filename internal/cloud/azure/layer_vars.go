package azure

import (
	"context"
	"fmt"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

// PrepareLayerVars sets resource_group_create (and state-backend RG name)
// so bootstrap and core-network do not fight over the same resource group.
func (p *Provider) PrepareLayerVars(ctx context.Context, spec cloud.StateSpec, layer stacks.Layer, merged map[string]any) error {
	switch layer {
	case stacks.LayerStateBackend, stacks.LayerCoreNetwork:
	default:
		return nil
	}
	rgName := SharedResourceGroupName(spec)
	exists, err := p.ResourceGroupExists(ctx, spec, rgName)
	if err != nil {
		return fmt.Errorf("check resource group %q: %w", rgName, err)
	}
	merged["resource_group_create"] = !exists
	if layer == stacks.LayerStateBackend {
		merged["resource_group_name"] = rgName
	}
	return nil
}