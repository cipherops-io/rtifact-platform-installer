package config

import "fmt"

// SchemaV1RetiredMessage includes a pasteable features block for the config cloud.
func SchemaV1RetiredMessage(cloud string) string {
	block := v1FeaturesBlock(cloud)
	return fmt.Sprintf(`%s

Paste this features block (tune for your environment):

%s`, SchemaV1RetiredHint, block)
}

func v1FeaturesBlock(cloud string) string {
	switch cloud {
	case "aws":
		return `features:
  network: true
  bastion: true
  kubernetes_cluster: true
  karpenter: true
  public_dns: true
  certificates: true`
	case "gcp":
		return `features:
  resource_group: true
  virtual_network: true
  kubernetes_cluster: true`
	default:
		return `features:
  resource_group: true
  virtual_network: true
  key_vault: true
  key_vault_private_endpoint: true
  postgresql_flexible_server: true
  container_registry: true
  storage_account: true
  kubernetes_cluster: true
  azure_bastion: true
  private_dns_zones: true`
	}
}