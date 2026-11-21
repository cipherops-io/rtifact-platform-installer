package config_test

import (
	"path/filepath"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/features"
)

func TestFeatureRegistryLoad(t *testing.T) {
	root := filepath.Join("..", "..")
	reg, err := features.LoadProfileRegistry(filepath.Join(root, "profiles"), "aios-infra", "azure")
	if err != nil {
		t.Fatal(err)
	}
	if len(reg.Features) < 5 {
		t.Fatalf("expected features in registry, got %d", len(reg.Features))
	}
	merged := map[string]any{
		"features": map[string]any{
			"resource_group":     true,
			"virtual_network":    true,
			"kubernetes_cluster": true,
		},
	}
	if err := config.DriftFeaturesKeys(merged, reg); err != nil {
		t.Fatal(err)
	}
}
