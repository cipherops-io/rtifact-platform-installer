package preflight

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud/registry"
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
	defaults, err := config.LoadProfileDefaults(profilesDir, cfg.Profile, cfg.Cloud)
	if err != nil {
		return err
	}
	merged, err := config.MergeBuildFromPath(defaults, configPath)
	if err != nil {
		return err
	}
	merged["cloud"] = cfg.Cloud
	if err := quota.ValidateMerged(configPath, merged); err != nil {
		return err
	}
	if err := checkCIDR(merged); err != nil {
		return err
	}
	if c, err := registry.Get(cfg.Cloud); err == nil {
		if err := c.Preflight(context.Background(), merged); err != nil {
			return err
		}
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
	var cidrs []string
	for _, k := range []string{"vnet_cidr", "node_subnet_cidr", "pod_cidr", "service_cidr"} {
		if v, ok := netw[k].(string); ok && v != "" {
			cidrs = append(cidrs, v)
		}
	}
	for i := 0; i < len(cidrs); i++ {
		_, n1, err := net.ParseCIDR(cidrs[i])
		if err != nil {
			return fmt.Errorf("network.%s invalid CIDR", cidrs[i])
		}
		for j := i + 1; j < len(cidrs); j++ {
			_, n2, err := net.ParseCIDR(cidrs[j])
			if err != nil {
				continue
			}
			if n1.Contains(n2.IP) || n2.Contains(n1.IP) {
				return fmt.Errorf("CIDR overlap: %s and %s", cidrs[i], cidrs[j])
			}
		}
	}
	if v, ok := netw["kubernetes_version"].(string); ok && strings.HasPrefix(v, "1.30") {
		fmt.Fprintln(os.Stderr, "warn: Kubernetes 1.30 may be unsupported in some regions; preflight did not verify AKS API")
	}
	_ = context.Background()
	return nil
}
