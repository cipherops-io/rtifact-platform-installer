package config

import (
	"os"
	"path/filepath"
	"strings"
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
schema_version: 1
cloud: azure
profile: ai-os-k8s
account:
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
schema_version: 1
cloud: azure
profile: ai-os-k8s
account:
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
func writeCfg(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

const v2Base = `
schema_version: 1
cloud: azure
profile: ai-os-k8s
account:
  subscription_id: "11111111-1111-1111-1111-111111111111"
  region: centralindia
workload:
  name_prefix: demo
`

func TestLoadCustomerSchemaV2(t *testing.T) {
	cfg, err := LoadCustomer(writeCfg(t, v2Base))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Cloud != "azure" || cfg.Account.Region != "centralindia" || cfg.Account.Environment != "public" {
		t.Fatalf("cfg = %+v", cfg)
	}
	if !cfg.IsAirGapped() {
		t.Fatal("state.air_gapped must default to true")
	}
	if !cfg.UsesLocalState() {
		t.Fatal("no state block means local state")
	}
}

func TestLoadCustomerRejectsUnsupportedSchemaVersion(t *testing.T) {
	wrong := `
schema_version: 2
cloud: azure
profile: aios-infra
account:
  subscription_id: "11111111-1111-1111-1111-111111111111"
  region: centralindia
workload:
  name_prefix: demo
`
	if _, err := LoadCustomer(writeCfg(t, wrong)); err == nil {
		t.Fatal("expected schema_version error")
	}
	noCloud := `
schema_version: 1
profile: aios-infra
account:
  subscription_id: "11111111-1111-1111-1111-111111111111"
  region: centralindia
workload:
  name_prefix: demo
`
	if _, err := LoadCustomer(writeCfg(t, noCloud)); err == nil {
		t.Fatal("expected cloud error")
	}
}

func TestAirGappedExplicitFalse(t *testing.T) {
	cfg, err := LoadCustomer(writeCfg(t, v2Base+"state:\n  backend: azurerm\n  air_gapped: false\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.IsAirGapped() || cfg.UsesLocalState() {
		t.Fatalf("air_gapped=%v local=%v", cfg.IsAirGapped(), cfg.UsesLocalState())
	}
}

func TestValidateCustomerYAMLFeatureRegistry(t *testing.T) {
	ok := v2Base + "features:\n  resource_group: true\n  virtual_network: true\n"
	if err := ValidateCustomerYAML(writeCfg(t, ok)); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
	unknown := v2Base + "features:\n  resource_group: true\n  made_up_feature: true\n"
	if err := ValidateCustomerYAML(writeCfg(t, unknown)); err == nil || !strings.Contains(err.Error(), "made_up_feature") {
		t.Fatalf("unknown feature not rejected: %v", err)
	}
	unsupported := v2Base + "features:\n  resource_group: true\n  virtual_network: true\n  mongodb_atlas: true\n"
	if err := ValidateCustomerYAML(writeCfg(t, unsupported)); err == nil {
		t.Fatal("unsupported feature not rejected")
	}
	missingDep := v2Base + "features:\n  virtual_network: false\n  kubernetes_cluster: true\n"
	if err := ValidateCustomerYAML(writeCfg(t, missingDep)); err == nil {
		t.Fatal("missing requires not rejected")
	}
}

func TestExamplesValidate(t *testing.T) {
	for _, p := range []string{"sandbox.yaml", "ai-os-centralindia/config.yaml"} {
		if err := ValidateCustomerYAML(filepath.Join("..", "..", "examples", p)); err != nil {
			t.Fatalf("%s: %v", p, err)
		}
	}
}
