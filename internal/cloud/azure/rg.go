package azure

import (
	"context"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud"
)

// WorkloadResourceGroupSuffix is the shared RG for workload + tfstate (modules/azure/naming).
const WorkloadResourceGroupSuffix = "-rg"

// SharedResourceGroupName returns the resource group used for workload and remote state.
func SharedResourceGroupName(spec cloud.StateSpec) string {
	if o := spec.Override; o != nil && o.ResourceGroupName != "" {
		return o.ResourceGroupName
	}
	return spec.NamePrefix + WorkloadResourceGroupSuffix
}

// StateResourceGroupName matches SharedResourceGroupName (legacy name kept for callers).
func StateResourceGroupName(namePrefix string) string {
	return namePrefix + WorkloadResourceGroupSuffix
}

// ResourceGroupExists reports whether name exists in the subscription (Azure CLI).
func (p *Provider) ResourceGroupExists(ctx context.Context, spec cloud.StateSpec, name string) (bool, error) {
	if strings.TrimSpace(spec.Account.SubscriptionID) == "" {
		return false, nil
	}
	if _, err := p.lookPath(AzureCLI); err != nil {
		return false, err
	}
	out, err := p.az(ctx, "group", "exists", "--name", name, "--subscription", spec.Account.SubscriptionID)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) == "true", nil
}
