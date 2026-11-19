package preflight

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/quota"
)

func Run(configPath string) error {
	if err := config.ValidateCustomerYAML(configPath); err != nil {
		return err
	}
	cfg, err := config.LoadCustomer(configPath)
	if err != nil {
		return err
	}
	root, _, err := assetsRoot()
	if err != nil {
		return err
	}
	profilesDir := root + "/profiles"
	defaults, err := config.LoadProfileDefaults(profilesDir, cfg.Profile)
	if err != nil {
		return err
	}
	merged, err := config.MergeBuildFromPath(defaults, configPath)
	if err != nil {
		return err
	}
	if err := quota.ValidateMerged(configPath, merged); err != nil {
		return err
	}
	if err := checkCIDR(merged); err != nil {
		return err
	}
	if cfg.State != nil && cfg.State.Backend == "azurerm" {
		fmt.Fprintln(os.Stderr, "preflight: remote state configured (reachability not checked)")
	}
	fmt.Fprintln(os.Stderr, "preflight: OK")
	return nil
}

func assetsRoot() (string, bool, error) {
	return assets.ContentRoot()
}

func checkCIDR(merged map[string]any) error {
	netw, ok := merged["network"].(map[string]any)
	if !ok {
		return nil
	}
	vnetCIDR, _ := netw["vnet_cidr"].(string)
	subnetKeys := []string{
		"node_subnet_cidr", "load_balancer_subnet_cidr", "private_endpoint_subnet_cidr",
		"management_subnet_cidr", "azure_firewall_subnet_cidr", "azure_bastion_subnet_cidr",
		"gateway_subnet_cidr", "dns_resolver_inbound_subnet_cidr", "dns_resolver_outbound_subnet_cidr",
	}
	var subnets []string
	for _, k := range subnetKeys {
		if v, ok := netw[k].(string); ok && v != "" {
			subnets = append(subnets, v)
		}
	}
	var overlays []string
	for _, k := range []string{"pod_cidr", "service_cidr"} {
		if v, ok := netw[k].(string); ok && v != "" {
			overlays = append(overlays, v)
		}
	}
	if vnetCIDR != "" {
		if _, _, err := net.ParseCIDR(vnetCIDR); err != nil {
			return fmt.Errorf("network.vnet_cidr invalid CIDR")
		}
	}
	var parsedSubnets []*net.IPNet
	for _, c := range subnets {
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			return fmt.Errorf("network CIDR %q invalid", c)
		}
		if vnetCIDR != "" {
			_, vn, err := net.ParseCIDR(vnetCIDR)
			if err == nil && !vn.Contains(n.IP) {
				return fmt.Errorf("subnet %s is not inside vnet %s", c, vnetCIDR)
			}
		}
		parsedSubnets = append(parsedSubnets, n)
	}
	for i := 0; i < len(parsedSubnets); i++ {
		for j := i + 1; j < len(parsedSubnets); j++ {
			if cidrRangesOverlap(parsedSubnets[i], parsedSubnets[j]) {
				return fmt.Errorf("CIDR overlap: %s and %s", subnets[i], subnets[j])
			}
		}
	}
	var parsedOverlays []*net.IPNet
	for _, c := range overlays {
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			return fmt.Errorf("network CIDR %q invalid", c)
		}
		parsedOverlays = append(parsedOverlays, n)
	}
	for i := 0; i < len(parsedOverlays); i++ {
		for j := i + 1; j < len(parsedOverlays); j++ {
			if cidrRangesOverlap(parsedOverlays[i], parsedOverlays[j]) {
				return fmt.Errorf("CIDR overlap: %s and %s", overlays[i], overlays[j])
			}
		}
	}
	if v, ok := netw["kubernetes_version"].(string); ok && strings.HasPrefix(v, "1.30") {
		fmt.Fprintln(os.Stderr, "warn: Kubernetes 1.30 may be unsupported in some regions; preflight did not verify AKS API")
	}
	_ = context.Background()
	return nil
}

func cidrRangesOverlap(a, b *net.IPNet) bool {
	return a.Contains(b.IP) || b.Contains(a.IP)
}
