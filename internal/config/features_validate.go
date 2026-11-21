package config

import (
	"fmt"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
	"github.com/cipherops-io/rtifact-platform-installer/internal/features"
	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
)

// validateMergedFeaturesFromRegistry validates merged features (a free-form map)
// against the profile feature registry: unknown keys, unsupported features and
// `requires` dependencies.
func validateMergedFeaturesFromRegistry(merged map[string]any, profileID, cloud string) error {
	if _, ok := merged["features"].(map[string]any); !ok {
		return nil
	}
	root, _, err := assets.ContentRootOptional()
	if err != nil {
		return fmt.Errorf("locate profiles for feature registry: %w", err)
	}
	reg, err := features.LoadProfileRegistry(filepath.Join(root, "profiles"), profileID, cloud)
	if err != nil {
		return fmt.Errorf("load feature registry for profile %q: %w", profileID, err)
	}
	feat := featuresBoolMap(merged)
	if err := reg.ValidateEnabled(feat); err != nil {
		return usererr.Wrap(err,
			"Feature flags are invalid for this profile.",
			"See profiles/"+profileID+"/features.yaml for allowed features, dependencies, and unsupported items.")
	}
	if err := validateFeatureBusinessRules(merged, feat); err != nil {
		return usererr.Wrap(err,
			"Feature combination is not allowed.",
			"Adjust features in your config or see profiles/"+profileID+"/features.yaml.")
	}
	return nil
}

func validateFeatureBusinessRules(merged map[string]any, feat map[string]bool) error {
	if feat["redis_cache"] && feat["azure_managed_redis"] {
		return fmt.Errorf("enable at most one of features.redis_cache and features.azure_managed_redis")
	}
	if feat["azure_vpn_gateway"] && feat["azure_express_route_gateway"] {
		return fmt.Errorf("enable at most one of features.azure_vpn_gateway and features.azure_express_route_gateway")
	}
	return nil
}

func validateAtLeastOneFeature(merged map[string]any) error {
	feat := featuresBoolMap(merged)
	for _, on := range feat {
		if on {
			return nil
		}
	}
	return fmt.Errorf("nothing to deploy: enable at least one feature in your config")
}

func featuresBoolMap(merged map[string]any) map[string]bool {
	out := map[string]bool{}
	featuresMap, ok := merged["features"].(map[string]any)
	if !ok {
		return out
	}
	for k, v := range featuresMap {
		b, _ := v.(bool)
		out[k] = b
	}
	return out
}

// DriftFeaturesKeys compares registry keys to merged feature keys (test helper).
func DriftFeaturesKeys(merged map[string]any, reg *features.Registry) error {
	if reg == nil {
		return nil
	}
	for k := range featuresBoolMap(merged) {
		if _, ok := reg.Features[k]; !ok {
			return fmt.Errorf("unknown feature key in merge: %s", k)
		}
	}
	return nil
}
