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
	def, err := LoadProfileDefaults(filepath.Join(root, "profiles"), cfg.Profile)
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

func WriteInitConfig(tier, customer, out string) error {
	body := fmt.Sprintf(`schema_version: 1
cloud: azure
profile: ai-os-k8s
tier: %s
customer:
  name: %s
  code: %s
account:
  subscription_id: "00000000-0000-0000-0000-000000000000"
  tenant_id: "00000000-0000-0000-0000-000000000000"
  region: centralindia
workload:
  name: aios
  environment: sbx
network:
  vnet_cidr: "10.0.0.0/20"
  node_subnet_cidr: "10.0.0.0/22"
  load_balancer_subnet_cidr: "10.0.4.0/24"
  private_endpoint_subnet_cidr: "10.0.5.0/24"
  management_subnet_cidr: "10.0.6.0/27"
  pod_cidr: "10.244.0.0/16"
  service_cidr: "172.20.0.0/16"
  dns_service_ip: "172.20.0.10"
state:
  backend: azurerm
  bootstrap: true
  air_gapped: true
`, tier, customer, customer)
	return os.WriteFile(out, []byte(body), 0o644)
}