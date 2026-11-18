package config

import "fmt"

// CustomerConfig is the external customer-supplied YAML contract (v1).
type CustomerConfig struct {
	SchemaVersion           int                         `yaml:"schema_version"`
	Profile                 string                      `yaml:"profile"`
	Tier                    string                      `yaml:"tier"`
	Customer                CustomerIdentity            `yaml:"customer,omitempty"`
	State                   *StateBlock                   `yaml:"state,omitempty"`
	Azure                   AzureBlock                  `yaml:"azure"`
	Workload                WorkloadBlock               `yaml:"workload"`
	Overrides               map[string]any              `yaml:"overrides,omitempty"`
	Features                *FeaturesBlock              `yaml:"features,omitempty"`
	Network                 map[string]any              `yaml:"network,omitempty"`
	KubernetesCluster       map[string]any              `yaml:"kubernetes_cluster,omitempty"`
	KeyVault                map[string]any              `yaml:"key_vault,omitempty"`
	PostgreSQLFlexibleServer map[string]any             `yaml:"postgresql_flexible_server,omitempty"`
	ContainerRegistry       map[string]any              `yaml:"container_registry,omitempty"`
	StorageAccount          map[string]any              `yaml:"storage_account,omitempty"`
	RedisCache                map[string]any              `yaml:"redis_cache,omitempty"`
	AzureBastion              map[string]any              `yaml:"azure_bastion,omitempty"`
	PrivateDNS                map[string]any              `yaml:"private_dns,omitempty"`
	Backend                   *BackendBlock               `yaml:"backend,omitempty"`
}

type AzureBlock struct {
	SubscriptionID string    `yaml:"subscription_id"`
	TenantID       string    `yaml:"tenant_id,omitempty"`
	Region         string    `yaml:"region"`
	Environment    string    `yaml:"environment,omitempty"`
	Auth           AuthBlock `yaml:"auth,omitempty"`
}

type AuthBlock struct {
	UseMSI bool `yaml:"use_msi"`
}

type CustomerIdentity struct {
	Name string `yaml:"name"`
	Code string `yaml:"code"`
}

type StateBlock struct {
	Backend   string `yaml:"backend"`
	Bootstrap bool   `yaml:"bootstrap"`
}

type WorkloadBlock struct {
	Name        string `yaml:"name"`
	NamePrefix  string `yaml:"name_prefix"`
	Environment string `yaml:"environment,omitempty"`
}

// FeaturesBlock gates stack modules; JSON/tfvars use the same snake_case keys.
type FeaturesBlock struct {
	ResourceGroup            bool `yaml:"resource_group"`
	VirtualNetwork           bool `yaml:"virtual_network"`
	KeyVault                 bool `yaml:"key_vault"`
	KeyVaultPrivateEndpoint  bool `yaml:"key_vault_private_endpoint"`
	PostgreSQLFlexibleServer bool `yaml:"postgresql_flexible_server"`
	ContainerRegistry        bool `yaml:"container_registry"`
	StorageAccount           bool `yaml:"storage_account"`
	RedisCache               bool `yaml:"redis_cache"`
	KubernetesCluster        bool `yaml:"kubernetes_cluster"`
	AzureBastion             bool `yaml:"azure_bastion"`
	PrivateDNSZones          bool `yaml:"private_dns_zones"`
}

type BackendBlock struct {
	Type               string `yaml:"type"`
	ResourceGroupName  string `yaml:"resource_group_name"`
	StorageAccountName string `yaml:"storage_account_name"`
	ContainerName      string `yaml:"container_name"`
	Key                string `yaml:"key"`
}

// NamePrefix returns the workload prefix used in Azure names and state keys.
func (c *CustomerConfig) NamePrefix() string {
	if c.Workload.NamePrefix != "" {
		return c.Workload.NamePrefix
	}
	if c.Customer.Code != "" && c.Workload.Name != "" && c.Workload.Environment != "" {
		return fmt.Sprintf("%s-%s-%s", c.Customer.Code, c.Workload.Name, c.Workload.Environment)
	}
	return c.Workload.NamePrefix
}

// ProfileMeta is profiles/<id>/profile.yaml.
type ProfileMeta struct {
	StackID     string `yaml:"stack_id"`
	Version     string `yaml:"version"`
	ChecklistID string `yaml:"checklist_id"`
}