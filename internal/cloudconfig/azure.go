package cloudconfig

import "context"

func init() {
	Register("azure", azureConfig{})
}

type azureConfig struct{}

func (azureConfig) TFVarSections() []string {
	return []string{
		"features", "state", "tags", "network", "kubernetes_cluster",
		"key_vault", "postgresql_flexible_server", "container_registry",
		"storage_account", "redis_cache", "azure_managed_redis",
		"azure_bastion", "bastion", "azure_firewall", "azure_vpn",
		"azure_express_route", "private_dns_resolver", "public_dns",
		"application_gateway", "disk_encryption_set", "ai_foundry",
		"private_dns", "azure_files", "communication_services",
	}
}

func (azureConfig) Normalize(merged map[string]any) {
	bastion, _ := merged["bastion"].(map[string]any)
	if bastion == nil {
		bastion = make(map[string]any)
	}
	if legacy, ok := merged["azure_bastion"].(map[string]any); ok {
		for _, k := range []string{"vm_size", "os_disk_size_gb", "os_disk_type"} {
			if _, set := bastion[k]; !set {
				if v, ok := legacy[k]; ok {
					bastion[k] = v
				}
			}
		}
	}
	if _, ok := bastion["managed"]; !ok {
		bastion["managed"] = false
	}
	if _, ok := bastion["os_disk_type"]; !ok {
		bastion["os_disk_type"] = "Premium_LRS"
	}
	if _, ok := bastion["vm_size"]; !ok {
		bastion["vm_size"] = "Standard_D4s_v5"
	}
	if _, ok := bastion["os_disk_size_gb"]; !ok {
		bastion["os_disk_size_gb"] = float64(100)
	}
	merged["bastion"] = bastion
}

func (azureConfig) Validate(_ map[string]any) error { return nil }

func (azureConfig) Preflight(context.Context, map[string]any) error { return nil }