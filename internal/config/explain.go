package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
)

func PrintExplain(customerPath string) error {
	root, _, err := assets.ContentRoot()
	if err != nil {
		return err
	}
	cfg, err := LoadCustomer(customerPath)
	if err != nil {
		return err
	}
	def, err := LoadProfileDefaults(filepath.Join(root, "profiles"), cfg.Profile, cfg.Cloud)
	if err != nil {
		return err
	}
	merged, err := MergeBuildFromPath(def, customerPath)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "profile defaults + tier %q + customer overrides\n", cfg.Tier)
	if kc, ok := merged["kubernetes_cluster"].(map[string]any); ok {
		if np, ok := kc["node_pools"].(map[string]any); ok {
			for name, v := range np {
				pm, _ := v.(map[string]any)
				fmt.Fprintf(os.Stdout, "kubernetes_cluster.node_pools.%s.vm_size = %v\n", name, pm["vm_size"])
				fmt.Fprintf(os.Stdout, "kubernetes_cluster.node_pools.%s.node_count = %v\n", name, pm["node_count"])
			}
		}
	}
	return nil
}

