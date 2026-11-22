package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud/registry"
	"github.com/cipherops-io/rtifact-platform-installer/internal/maputil"
	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
	"gopkg.in/yaml.v3"
)

// LoadCustomer reads and parses customer config.yaml.
func LoadCustomer(path string) (*CustomerConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg CustomerConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse customer config: %w", err)
	}
	if cfg.SchemaVersion != SchemaVersion {
		return nil, fmt.Errorf("schema_version must be %d (got %d)", SchemaVersion, cfg.SchemaVersion)
	}
	if cfg.Cloud == "" {
		return nil, fmt.Errorf("cloud is required (supported: %s)", strings.Join(registry.Supported, ", "))
	}
	if _, err := registry.Get(cfg.Cloud); err != nil {
		return nil, err
	}
	if cfg.Profile == "" {
		return nil, fmt.Errorf("profile is required")
	}
	cfg.Profile = ResolveProfileID(cfg.Profile)
	usererr.SetProfileID(cfg.Profile)
	if err := validateAccountForCloud(&cfg); err != nil {
		return nil, err
	}
	if cfg.Account.Region == "" {
		return nil, fmt.Errorf("account.region is required")
	}
	if cfg.NamePrefix() == "" {
		return nil, fmt.Errorf("workload.name_prefix or customer.code+workload.name+workload.environment is required")
	}
	if cfg.Account.Environment == "" {
		cfg.Account.Environment = "public"
	}
	return &cfg, nil
}

// LoadProfileDefaults reads profiles/<id>/<cloud>/defaults.yaml (or legacy profiles/<id>/defaults.yaml).
func LoadProfileDefaults(profilesDir, profileID, cloud string) (map[string]any, error) {
	profileID = ResolveProfileID(profileID)
	path := filepath.Join(profilesDir, profileID, cloud, "defaults.yaml")
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			path = filepath.Join(profilesDir, profileID, "defaults.yaml")
			raw, err = os.ReadFile(path)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("profile defaults %q/%s: %w", profileID, cloud, err)
	}
	var defaults map[string]any
	if err := yaml.Unmarshal(raw, &defaults); err != nil {
		return nil, err
	}
	return defaults, nil
}

// LoadProfileMeta reads profiles/<id>/profile.yaml from work dir.
func LoadProfileMeta(profilesDir, profileID string) (*ProfileMeta, error) {
	profileID = ResolveProfileID(profileID)
	path := filepath.Join(profilesDir, profileID, "profile.yaml")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var meta ProfileMeta
	if err := yaml.Unmarshal(raw, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

var customerTFVarSections = []string{
	"features",
	"state",
	"tags",
	"certificates",
	"network",
	"kubernetes_cluster",
	"key_vault",
	"postgresql_flexible_server",
	"container_registry",
	"storage_account",
	"redis_cache",
	"azure_managed_redis",
	"azure_bastion",
	"bastion",
	"azure_firewall",
	"azure_vpn",
	"azure_express_route",
	"private_dns_resolver",
	"public_dns",
	"application_gateway",
	"disk_encryption_set",
	"ai_foundry",
	"private_dns",
	"azure_files",
	"communication_services",
}

var replaceMergeSections = map[string][]string{
	"kubernetes_cluster": {"node_pools"},
}

func LoadProfileTier(profilesDir, profileID, cloud, tier string) (map[string]any, error) {
	if tier == "" {
		return nil, nil
	}
	path := filepath.Join(profilesDir, profileID, cloud, "tiers", tier+".yaml")
	if _, err := os.Stat(path); err != nil {
		path = filepath.Join(profilesDir, profileID, "tiers", tier+".yaml")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("tier %q: %w", tier, err)
	}
	var tierMap map[string]any
	if err := yaml.Unmarshal(raw, &tierMap); err != nil {
		return nil, err
	}
	return normalizeAnyMap(tierMap), nil
}

// MergeBuildFromPath merges profile defaults, tier, and customer YAML into OpenTofu variables.
func MergeBuildFromPath(defaults map[string]any, customerPath string) (map[string]any, error) {
	customer, err := LoadCustomer(customerPath)
	if err != nil {
		return nil, err
	}
	c, err := registry.Get(customer.Cloud)
	if err != nil {
		return nil, err
	}
	sections := c.TFVarSections()
	rawBytes, err := os.ReadFile(customerPath)
	if err != nil {
		return nil, err
	}
	var raw map[string]any
	if err := yaml.Unmarshal(rawBytes, &raw); err != nil {
		return nil, err
	}
	raw = normalizeAnyMap(raw)

	if err := validateCustomerExplicitFeatures(raw, customer.Profile, customer.Cloud); err != nil {
		return nil, err
	}

	contentRoot, _, _ := assets.ContentRootOptional()
	var tierMap map[string]any
	if customer.Tier != "" && contentRoot != "" {
		tierMap, err = LoadProfileTier(filepath.Join(contentRoot, "profiles"), customer.Profile, customer.Cloud, customer.Tier)
		if err != nil {
			return nil, err
		}
	}

	merged := maputil.DeepCopyMap(defaults)
	if tierMap != nil {
		mergeSectionsInto(merged, tierMap, sections)
	}
	if customer.Overrides != nil {
		mergeSectionsInto(merged, customer.Overrides, sections)
	}
	applyCustomerScalars(merged, customer)

	for _, section := range sections {
		val, ok := raw[section]
		if !ok || val == nil {
			continue
		}
		src, ok := toStringMap(val)
		if !ok {
			return nil, fmt.Errorf("section %q must be a mapping", section)
		}
		if childKeys, isReplace := replaceMergeSections[section]; isReplace {
			if childKeys == nil {
				merged[section] = maputil.DeepCopyMap(src)
				continue
			}
			existing, _ := merged[section].(map[string]any)
			if existing == nil {
				existing = make(map[string]any)
			} else {
				existing = maputil.DeepCopyMap(existing)
			}
			for _, child := range childKeys {
				if v, ok := src[child]; ok {
					existing[child] = v
					delete(src, child)
				}
			}
			deepMerge(existing, src)
			merged[section] = existing
			continue
		}
		if existing, ok := merged[section].(map[string]any); ok {
			deepMerge(existing, src)
		} else {
			merged[section] = maputil.DeepCopyMap(src)
		}
	}

	c.Normalize(merged)
	if err := validateMergedFeaturesFromRegistry(merged, customer.Profile, customer.Cloud); err != nil {
		return nil, err
	}
	if err := c.Validate(merged); err != nil {
		return nil, err
	}
	return merged, nil
}

func mergeSectionsInto(merged, src map[string]any, sections []string) {
	for _, section := range sections {
		val, ok := src[section]
		if !ok || val == nil {
			continue
		}
		sm, ok := toStringMap(val)
		if !ok {
			continue
		}
		if childKeys, isReplace := replaceMergeSections[section]; isReplace {
			existing, _ := merged[section].(map[string]any)
			if existing == nil {
				existing = make(map[string]any)
			} else {
				existing = maputil.DeepCopyMap(existing)
			}
			for _, child := range childKeys {
				if v, ok := sm[child]; ok {
					existing[child] = v
					delete(sm, child)
				}
			}
			deepMerge(existing, sm)
			merged[section] = existing
			continue
		}
		if existing, ok := merged[section].(map[string]any); ok {
			deepMerge(existing, sm)
		} else {
			merged[section] = maputil.DeepCopyMap(sm)
		}
	}
}

func applyCustomerScalars(merged map[string]any, customer *CustomerConfig) {
	merged["location"] = customer.Account.Region
	merged["name_prefix"] = customer.NamePrefix()
	switch customer.Cloud {
	case "azure":
		merged["subscription_id"] = customer.Account.SubscriptionID
		merged["azure_environment"] = customer.Account.Environment
		merged["use_msi"] = customer.Account.Auth.UseMSI
	case "aws":
		merged["account_id"] = customer.Account.AccountID
	case "gcp":
		merged["project_id"] = customer.Account.ProjectID
	}
	if customer.Account.TenantID != "" {
		merged["tenant_id"] = customer.Account.TenantID
	}
	if customer.Workload.Environment != "" {
		merged["workload_environment"] = customer.Workload.Environment
	}
}

func validateAccountForCloud(cfg *CustomerConfig) error {
	switch cfg.Cloud {
	case "azure":
		if cfg.Account.SubscriptionID == "" {
			return fmt.Errorf("account.subscription_id is required for cloud azure")
		}
	case "aws":
		if cfg.Account.AccountID == "" {
			return fmt.Errorf("account.account_id is required for cloud aws")
		}
	case "gcp":
		if cfg.Account.ProjectID == "" {
			return fmt.Errorf("account.project_id is required for cloud gcp")
		}
	default:
		return fmt.Errorf("unsupported cloud %q", cfg.Cloud)
	}
	return nil
}

func normalizeAnyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = normalizeValue(v)
	}
	return out
}

func normalizeValue(v any) any {
	switch t := v.(type) {
	case map[interface{}]interface{}:
		return maputil.YAMLMapToAnyMap(t)
	case map[string]any:
		m := make(map[string]any, len(t))
		for k, val := range t {
			m[k] = normalizeValue(val)
		}
		return m
	case []interface{}:
		s := make([]any, len(t))
		for i, val := range t {
			s[i] = normalizeValue(val)
		}
		return s
	default:
		return v
	}
}

func toStringMap(v any) (map[string]any, bool) {
	switch t := v.(type) {
	case map[string]any:
		return t, true
	case map[interface{}]interface{}:
		return maputil.YAMLMapToAnyMap(t), true
	default:
		return nil, false
	}
}


func deepMerge(dst, src map[string]any) {
	for k, v := range src {
		if vMap, ok := v.(map[string]any); ok {
			if existing, ok := dst[k].(map[string]any); ok {
				deepMerge(existing, vMap)
				continue
			}
		}
		dst[k] = v
	}
}

// InstallerVarFile is passed to tofu plan/apply and auto-loaded if present.
const InstallerVarFile = "installer.auto.tfvars.json"

// WriteAutoTofuVarsJSON writes merged customer variables for the stack root module.
func WriteAutoTofuVarsJSON(stackDir string, vars map[string]any) error {
	b, err := json.MarshalIndent(vars, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(stackDir, InstallerVarFile), b, 0o644)
}
