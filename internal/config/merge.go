package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
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
	if cfg.Profile == "" {
		return nil, fmt.Errorf("profile is required")
	}
	if cfg.Azure.SubscriptionID == "" {
		return nil, fmt.Errorf("azure.subscription_id is required")
	}
	if cfg.Azure.Region == "" {
		return nil, fmt.Errorf("azure.region is required")
	}
	if cfg.NamePrefix() == "" {
		return nil, fmt.Errorf("workload.name_prefix or customer.code+workload.name+workload.environment is required")
	}
	if cfg.Azure.Environment == "" {
		cfg.Azure.Environment = "public"
	}
	return &cfg, nil
}

// LoadProfileDefaults reads profiles/<id>/defaults.yaml from embedded work dir.
func LoadProfileDefaults(profilesDir, profileID string) (map[string]any, error) {
	path := filepath.Join(profilesDir, profileID, "defaults.yaml")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("profile defaults %q: %w", profileID, err)
	}
	var defaults map[string]any
	if err := yaml.Unmarshal(raw, &defaults); err != nil {
		return nil, err
	}
	return defaults, nil
}

// LoadProfileMeta reads profiles/<id>/profile.yaml from work dir.
func LoadProfileMeta(profilesDir, profileID string) (*ProfileMeta, error) {
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
	"network",
	"kubernetes_cluster",
	"key_vault",
	"postgresql_flexible_server",
	"container_registry",
	"storage_account",
	"redis_cache",
	"azure_bastion",
	"private_dns",
}

var replaceMergeSections = map[string][]string{
	"kubernetes_cluster": {"node_pools"},
}

func LoadProfileTier(profilesDir, profileID, tier string) (map[string]any, error) {
	if tier == "" {
		return nil, nil
	}
	path := filepath.Join(profilesDir, profileID, "tiers", tier+".yaml")
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
	rawBytes, err := os.ReadFile(customerPath)
	if err != nil {
		return nil, err
	}
	var raw map[string]any
	if err := yaml.Unmarshal(rawBytes, &raw); err != nil {
		return nil, err
	}
	raw = normalizeAnyMap(raw)

	contentRoot, _, _ := assets.ContentRootOptional()
	var tierMap map[string]any
	if customer.Tier != "" && contentRoot != "" {
		tierMap, err = LoadProfileTier(filepath.Join(contentRoot, "profiles"), customer.Profile, customer.Tier)
		if err != nil {
			return nil, err
		}
	}

	merged := deepCopyMap(defaults)
	if tierMap != nil {
		mergeSectionsInto(merged, tierMap)
	}
	if customer.Overrides != nil {
		mergeSectionsInto(merged, customer.Overrides)
	}
	applyCustomerScalars(merged, customer)

	for _, section := range customerTFVarSections {
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
				merged[section] = deepCopyMap(src)
				continue
			}
			existing, _ := merged[section].(map[string]any)
			if existing == nil {
				existing = make(map[string]any)
			} else {
				existing = deepCopyMap(existing)
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
			merged[section] = deepCopyMap(src)
		}
	}

	if err := validateMergedFeaturesFromRegistry(merged, customer.Profile); err != nil {
		return nil, err
	}
	return merged, nil
}

func mergeSectionsInto(merged, src map[string]any) {
	for _, section := range customerTFVarSections {
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
				existing = deepCopyMap(existing)
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
			merged[section] = deepCopyMap(sm)
		}
	}
}

func applyCustomerScalars(merged map[string]any, customer *CustomerConfig) {
	merged["subscription_id"] = customer.Azure.SubscriptionID
	merged["location"] = customer.Azure.Region
	merged["azure_environment"] = customer.Azure.Environment
	merged["use_msi"] = customer.Azure.Auth.UseMSI
	merged["name_prefix"] = customer.NamePrefix()
	if customer.Azure.TenantID != "" {
		merged["tenant_id"] = customer.Azure.TenantID
	}
	if customer.Workload.Environment != "" {
		merged["workload_environment"] = customer.Workload.Environment
	}
}

func validateMergedFeatures(merged map[string]any) error {
	features, ok := merged["features"].(map[string]any)
	if !ok {
		return nil
	}
	boolFeature := func(key string) bool {
		v, ok := features[key]
		if !ok {
			return false
		}
		b, _ := v.(bool)
		return b
	}
	if boolFeature("redis_cache") {
		return fmt.Errorf("features.redis_cache: legacy Azure Cache for Redis is not supported in IaC yet; keep redis_cache false")
	}
	needsVNet := []string{
		"key_vault_private_endpoint",
		"postgresql_flexible_server",
		"container_registry",
		"storage_account",
		"redis_cache",
		"kubernetes_cluster",
		"azure_bastion",
	}
	for _, feat := range needsVNet {
		if boolFeature(feat) && !boolFeature("virtual_network") {
			return fmt.Errorf("features.%s requires features.virtual_network true", feat)
		}
	}
	if boolFeature("key_vault_private_endpoint") && !boolFeature("key_vault") {
		return fmt.Errorf("features.key_vault_private_endpoint requires features.key_vault true")
	}
	if boolFeature("kubernetes_cluster") && !boolFeature("resource_group") {
		return fmt.Errorf("features.kubernetes_cluster requires features.resource_group true")
	}
	anyOther := false
	for k, v := range features {
		if k == "resource_group" {
			continue
		}
		if b, ok := v.(bool); ok && b {
			anyOther = true
			break
		}
	}
	if anyOther && !boolFeature("resource_group") {
		return fmt.Errorf("features.resource_group must be true when deploying any other feature")
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
		return yamlMapToAnyMap(t)
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
		return yamlMapToAnyMap(t), true
	default:
		return nil, false
	}
}

func deepCopyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		switch t := v.(type) {
		case map[string]any:
			out[k] = deepCopyMap(t)
		case map[interface{}]interface{}:
			out[k] = yamlMapToAnyMap(t)
		default:
			out[k] = v
		}
	}
	return out
}

func yamlMapToAnyMap(in map[interface{}]interface{}) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		key, _ := k.(string)
		switch t := v.(type) {
		case map[interface{}]interface{}:
			out[key] = yamlMapToAnyMap(t)
		default:
			out[key] = normalizeValue(v)
		}
	}
	return out
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