package gcp

import "context"

func (*Provider) TFVarSections() []string {
	return []string{
		"features", "state", "tags", "network", "kubernetes_cluster", "bastion",
	}
}

func (*Provider) Normalize(merged map[string]any) {
	merged["bastion"] = merged["bastion"]
}

func (*Provider) Validate(_ map[string]any) error { return nil }

func (*Provider) Preflight(context.Context, map[string]any) error { return nil }