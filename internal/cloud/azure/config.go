package azure

import (
	"context"
	"fmt"
	"net"

	"github.com/cipherops-io/rtifact-platform-installer/internal/maputil"
)

func (*Provider) ExpandCustomerPaths(string, map[string]any) error { return nil }

func (*Provider) TFVarSections() []string {
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

func (*Provider) Normalize(merged map[string]any) {
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
	if a, _ := bastion["access"].(string); a == "" {
		// Legacy bastion.managed: true meant Azure Bastion; it now fronts the bastion VM.
		if maputil.Truthy(bastion["managed"]) {
			bastion["access"] = "azure_bastion"
		} else {
			bastion["access"] = "private"
		}
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

func (*Provider) Validate(merged map[string]any) error {
	features, _ := merged["features"].(map[string]any)
	if !maputil.Truthy(features["azure_bastion"]) {
		return nil
	}
	bastion, _ := merged["bastion"].(map[string]any)
	switch access, _ := bastion["access"].(string); access {
	case "private", "public_ip":
	case "azure_bastion":
		netw, _ := merged["network"].(map[string]any)
		cidr, _ := netw["azure_bastion_subnet_cidr"].(string)
		if cidr == "" {
			return fmt.Errorf("bastion.access is azure_bastion: set network.azure_bastion_subnet_cidr (a /26 or larger AzureBastionSubnet)")
		}
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			return fmt.Errorf("network.azure_bastion_subnet_cidr: %w", err)
		}
		if ones, _ := ipnet.Mask.Size(); ones > 26 {
			return fmt.Errorf("network.azure_bastion_subnet_cidr %s is too small: Azure Bastion needs /26 or larger", cidr)
		}
	default:
		return fmt.Errorf("bastion.access must be private, public_ip or azure_bastion (got %q)", access)
	}
	if access, _ := bastion["access"].(string); access == "public_ip" {
		if cidrs, _ := maputil.StringSlice(bastion["allowed_ssh_cidrs"]); len(cidrs) == 0 {
			return fmt.Errorf("bastion.access is public_ip: set bastion.allowed_ssh_cidrs to the operator IPs (for example 203.0.113.10/32)")
		}
	}
	return nil
}

func (*Provider) Preflight(context.Context, map[string]any) error { return nil }