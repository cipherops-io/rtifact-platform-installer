package config

import (
	"fmt"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
	"github.com/cipherops-io/rtifact-platform-installer/internal/features"
	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
)

const explicitFeaturesSummary = "Nothing to deploy: this config enables no features."

func validateCustomerExplicitFeatures(raw map[string]any, profileID, cloud string) error {
	if customerHasEnabledFeature(raw) {
		return nil
	}
	root, _, err := assets.ContentRootOptional()
	if err != nil || root == "" {
		return fmt.Errorf("locate profiles for feature hint: %w", err)
	}
	reg, err := features.LoadProfileRegistry(filepath.Join(root, "profiles"), profileID, cloud)
	if err != nil {
		return fmt.Errorf("load feature registry: %w", err)
	}
	hint := reg.PasteableFeaturesYAML()
	return &usererr.Error{
		Summary: explicitFeaturesSummary,
		Hint:    "Add an explicit features block to your config YAML, for example:\n\n" + hint,
	}
}

func customerHasEnabledFeature(raw map[string]any) bool {
	featuresMap, ok := raw["features"].(map[string]any)
	if !ok || len(featuresMap) == 0 {
		return false
	}
	for _, v := range featuresMap {
		if b, ok := v.(bool); ok && b {
			return true
		}
	}
	return false
}