package config

import (
	"fmt"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
	"github.com/cipherops-io/rtifact-platform-installer/internal/features"
)

func validateMergedFeaturesFromRegistry(merged map[string]any, profileID string) error {
	featuresMap, ok := merged["features"].(map[string]any)
	if !ok {
		return nil
	}
	bools := make(map[string]bool, len(featuresMap))
	for k, v := range featuresMap {
		b, _ := v.(bool)
		bools[k] = b
	}
	root, _, err := assets.ContentRootOptional()
	if err == nil && root != "" {
		reg, err := features.LoadProfileRegistry(filepath.Join(root, "profiles"), profileID)
		if err == nil {
			if err := reg.ValidateEnabled(bools); err != nil {
				return err
			}
			return nil
		}
	}
	return validateMergedFeatures(merged)
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