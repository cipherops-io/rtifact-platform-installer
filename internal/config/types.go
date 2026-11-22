package config

import "fmt"

// SchemaVersion is the customer config schema version this build understands.
const SchemaVersion = 1

// CustomerConfig is the external customer-supplied YAML contract (schema v1).
type CustomerConfig struct {
	SchemaVersion            int              `yaml:"schema_version"`
	Cloud                    string           `yaml:"cloud"`
	Profile                  string           `yaml:"profile"`
	Tier                     string           `yaml:"tier"`
	Customer                 CustomerIdentity `yaml:"customer,omitempty"`
	State                    *StateBlock      `yaml:"state,omitempty"`
	Account                  AccountBlock     `yaml:"account"`
	Workload                 WorkloadBlock    `yaml:"workload"`
	Overrides                map[string]any   `yaml:"overrides,omitempty"`
	Features                 map[string]bool  `yaml:"features,omitempty"`
	Network                  map[string]any   `yaml:"network,omitempty"`
	KubernetesCluster        map[string]any   `yaml:"kubernetes_cluster,omitempty"`
	KeyVault                 map[string]any   `yaml:"key_vault,omitempty"`
	PostgreSQLFlexibleServer map[string]any   `yaml:"postgresql_flexible_server,omitempty"`
	ContainerRegistry        map[string]any   `yaml:"container_registry,omitempty"`
	StorageAccount           map[string]any   `yaml:"storage_account,omitempty"`
	RedisCache               map[string]any   `yaml:"redis_cache,omitempty"`
	AzureBastion             map[string]any   `yaml:"azure_bastion,omitempty"`
	Bastion                  map[string]any   `yaml:"bastion,omitempty"`
	PrivateDNS               map[string]any   `yaml:"private_dns,omitempty"`
	Backend                  *BackendBlock    `yaml:"backend,omitempty"`
}

// AccountBlock identifies the cloud account to deploy into.
type AccountBlock struct {
	SubscriptionID string    `yaml:"subscription_id,omitempty"`
	TenantID       string    `yaml:"tenant_id,omitempty"`
	AccountID      string    `yaml:"account_id,omitempty"`
	ProjectID      string    `yaml:"project_id,omitempty"`
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
	Backend    string `yaml:"backend"`
	Bootstrap  bool   `yaml:"bootstrap"`
	BucketName string `yaml:"bucket_name,omitempty"`
	Encryption string `yaml:"encryption,omitempty"`
	Versioning *bool  `yaml:"versioning,omitempty"`
	// AirGapped defaults to true when unset: only CLI tools, no cloud SDK calls.
	AirGapped *bool `yaml:"air_gapped,omitempty"`
}

type WorkloadBlock struct {
	Name        string `yaml:"name"`
	NamePrefix  string `yaml:"name_prefix"`
	Environment string `yaml:"environment,omitempty"`
}

// BackendBlock optionally points at a pre-created remote state backend.
// Key is used as a prefix; the layer name is always appended.
type BackendBlock struct {
	Type               string `yaml:"type"`
	ResourceGroupName  string `yaml:"resource_group_name,omitempty"`
	StorageAccountName string `yaml:"storage_account_name,omitempty"`
	ContainerName      string `yaml:"container_name,omitempty"`
	Key                string `yaml:"key,omitempty"`
	S3Bucket           string `yaml:"s3_bucket,omitempty"`
	DynamoDBTable      string `yaml:"dynamodb_table,omitempty"`
	S3Region           string `yaml:"s3_region,omitempty"`
	GCSBucket          string `yaml:"gcs_bucket,omitempty"`
	GCSPrefix          string `yaml:"gcs_prefix,omitempty"`
}

// IsAirGapped reports state.air_gapped (default true).
func (c *CustomerConfig) IsAirGapped() bool {
	if c.State == nil || c.State.AirGapped == nil {
		return true
	}
	return *c.State.AirGapped
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
	StackID       string            `yaml:"stack_id"`
	Version       string            `yaml:"version"`
	ChecklistID   string            `yaml:"checklist_id"`
	Clouds        []string          `yaml:"clouds,omitempty"`
	CloudMaturity map[string]string `yaml:"cloud_maturity,omitempty"`
}

// CloudMaturityFor returns operational, preview, or unknown.
func (m *ProfileMeta) CloudMaturityFor(cloud string) string {
	if m == nil || m.CloudMaturity == nil {
		if cloud == "azure" {
			return "operational"
		}
		return "preview"
	}
	if v, ok := m.CloudMaturity[cloud]; ok && v != "" {
		return v
	}
	if cloud == "azure" {
		return "operational"
	}
	return "preview"
}
