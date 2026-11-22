package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergeRejectsMissingFeaturesBlock(t *testing.T) {
	defaults := map[string]any{"features": map[string]any{"network": false}}
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	body := `
schema_version: 1
cloud: aws
profile: aios-infra
account:
  account_id: "123456789012"
  region: us-east-1
workload:
  name_prefix: demo
`
	if err := os.WriteFile(cfgPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := MergeBuildFromPath(defaults, cfgPath)
	if err == nil || !strings.Contains(err.Error(), explicitFeaturesSummary) {
		t.Fatalf("expected explicit features error, got %v", err)
	}
	if !strings.Contains(err.Error(), "features:") {
		t.Fatalf("expected paste hint with features block: %v", err)
	}
}

func TestMergeRejectsAllFeaturesFalse(t *testing.T) {
	defaults := map[string]any{"features": map[string]any{"network": false}}
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	body := `
schema_version: 1
cloud: aws
profile: aios-infra
account:
  account_id: "123456789012"
  region: us-east-1
workload:
  name_prefix: demo
features:
  network: false
  bastion: false
`
	if err := os.WriteFile(cfgPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := MergeBuildFromPath(defaults, cfgPath)
	if err == nil || !strings.Contains(err.Error(), explicitFeaturesSummary) {
		t.Fatalf("expected explicit features error, got %v", err)
	}
}