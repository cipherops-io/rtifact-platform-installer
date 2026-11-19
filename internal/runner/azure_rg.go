package runner

import (
	"context"
	"fmt"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud"
	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud/azure"
	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

// injectAzureResourceGroupFlags sets resource_group_create (and state-backend RG name)
// so bootstrap and core-network do not fight over the same resource group.
func injectAzureResourceGroupFlags(customer *config.CustomerConfig, provider cloud.Provider, spec cloud.StateSpec, layer stacks.Layer, merged map[string]any) error {
	if customer.Cloud != azure.CloudName {
		return nil
	}
	switch layer {
	case stacks.LayerStateBackend, stacks.LayerCoreNetwork:
	default:
		return nil
	}

	ap, ok := provider.(*azure.Provider)
	if !ok {
		return nil
	}
	rgName := azure.SharedResourceGroupName(spec)
	exists, err := ap.ResourceGroupExists(context.Background(), spec, rgName)
	if err != nil {
		return fmt.Errorf("check resource group %q: %w", rgName, err)
	}
	merged["resource_group_create"] = !exists
	if layer == stacks.LayerStateBackend {
		merged["resource_group_name"] = rgName
	}
	return nil
}