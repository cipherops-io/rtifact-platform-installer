package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMergeBuildCustomerFields(t *testing.T) {
	defaults := map[string]any{
		"network": map[string]any{"vnet_cidr": "10.0.0.0/20"},
		"features": map[string]any{
			"resource_group": true,
			"virtual_network": true,
			"redis_cache": false,
		},
	}
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	body := `
profile: ai-os-aks
azure:
  subscription_id: "11111111-1111-1111-1111-111111111111"
  region: centralindia
workload:
  name_prefix: demo
features:
  resource_group: true
  virtual_network: true
  key_vault: true
  key_vault_private_endpoint: false
  postgresql_flexible_server: false
  container_registry: false
  storage_account: false
  redis_cache: false
  kubernetes_cluster: false
  azure_bastion: false
  private_dns_zones: false
`
	if err := os.WriteFile(cfgPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	merged, err := MergeBuildFromPath(defaults, cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if merged["name_prefix"] != "demo" {
		t.Fatalf("name_prefix: %v", merged["name_prefix"])
	}
	if merged["location"] != "centralindia" {
		t.Fatalf("location: %v", merged["location"])
	}
	features, _ := merged["features"].(map[string]any)
	if features["key_vault"] != true {
		t.Fatalf("features.key_vault: %v", features["key_vault"])
	}
}

func TestMergeReplaceNodePools(t *testing.T) {
	defaults := map[string]any{
		"features": map[string]any{
			"resource_group": true, "virtual_network": true, "redis_cache": false,
			"kubernetes_cluster": true,
		},
		"kubernetes_cluster": map[string]any{
			"node_pools": map[string]any{
				"system": map[string]any{"name": "system", "vm_size": "Standard_D4s_v5", "node_count": 3},
				"user_microservices": map[string]any{"name": "appa", "vm_size": "Standard_D16s_v5", "node_count": 3},
			},
		},
	}
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	body := `
profile: ai-os-aks
azure:
  subscription_id: "11111111-1111-1111-1111-111111111111"
  region: centralindia
workload:
  name_prefix: demo
features:
  resource_group: true
  virtual_network: true
  key_vault: false
  key_vault_private_endpoint: false
  postgresql_flexible_server: false
  container_registry: false
  storage_account: false
  redis_cache: false
  kubernetes_cluster: true
  azure_bastion: false
  private_dns_zones: false
kubernetes_cluster:
  node_pools:
    system:
      name: system
      vm_size: Standard_D2s_v3
      node_count: 1
`
	if err := os.WriteFile(cfgPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	merged, err := MergeBuildFromPath(defaults, cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	kc, _ := merged["kubernetes_cluster"].(map[string]any)
	pools, _ := kc["node_pools"].(map[string]any)
	if _, ok := pools["user_microservices"]; ok {
		t.Fatal("expected user_microservices pool to be replaced away")
	}
	if pools["system"] == nil {
		t.Fatal("expected system pool")
	}
}