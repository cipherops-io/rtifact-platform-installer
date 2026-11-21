package cloudconfig

import "context"

func init() {
	Register("gcp", gcpConfig{})
}

type gcpConfig struct{}

func (gcpConfig) TFVarSections() []string {
	return []string{
		"features", "state", "tags", "network", "kubernetes_cluster", "bastion",
	}
}

func (gcpConfig) Normalize(merged map[string]any) {
	merged["bastion"] = merged["bastion"]
}

func (gcpConfig) Validate(_ map[string]any) error { return nil }

func (gcpConfig) Preflight(context.Context, map[string]any) error { return nil }