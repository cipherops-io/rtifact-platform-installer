package aws

import (
	"strings"
	"testing"
)

func TestValidatePodIdentitiesRequiresEntriesAndAddon(t *testing.T) {
	p := New()
	features := map[string]any{"pod_identities": true}
	kc := map[string]any{
		"addons": map[string]any{
			"vpc-cni": map[string]any{},
		},
	}
	merged := map[string]any{
		"features":           features,
		"kubernetes_cluster": kc,
	}
	if err := p.Validate(merged); err == nil || !strings.Contains(err.Error(), "at least one entry") {
		t.Fatalf("want empty list error, got %v", err)
	}

	kc["pod_identities"] = []any{
		map[string]any{
			"name":            "a",
			"namespace":       "ns",
			"service_account": "sa",
			"policy_arns":     []any{"arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess"},
		},
		map[string]any{
			"name":            "a",
			"namespace":       "ns2",
			"service_account": "sa2",
			"policy_arns":     []any{"arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess"},
		},
	}
	if err := p.Validate(merged); err == nil || !strings.Contains(err.Error(), "duplicate name") {
		t.Fatalf("want duplicate name error, got %v", err)
	}

	kc["pod_identities"] = []any{
		map[string]any{
			"name":            "external-dns",
			"namespace":       "external-dns",
			"service_account": "external-dns",
			"policy_arns":     []any{"arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess"},
		},
	}
	if err := p.Validate(merged); err == nil || !strings.Contains(err.Error(), "eks-pod-identity-agent") {
		t.Fatalf("want addon error, got %v", err)
	}

	kc["addons"].(map[string]any)["eks-pod-identity-agent"] = map[string]any{"version": "v1"}
	if err := p.Validate(merged); err != nil {
		t.Fatalf("valid pod identities: %v", err)
	}
}
