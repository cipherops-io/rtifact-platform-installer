package aws

import (
	"strings"
	"testing"
)

func TestValidateBastion(t *testing.T) {
	for name, tc := range map[string]struct {
		bastion map[string]any
		want    string
	}{
		"ssm needs no cidrs": {bastion: map[string]any{"access": "ssm"}},
		"ssh with cidrs":     {bastion: map[string]any{"access": "ssh", "allowed_cidrs": []any{"203.0.113.0/24"}}},
		"ssh without cidrs":  {bastion: map[string]any{"access": "ssh"}, want: "set bastion.allowed_cidrs"},
		"ssh open to all":    {bastion: map[string]any{"access": "ssh", "allowed_cidrs": []any{"0.0.0.0/0"}}, want: "opens SSH to the internet"},
		"ssh bad cidr":       {bastion: map[string]any{"access": "ssh", "allowed_cidrs": []any{"10.0.0.1"}}, want: "is not a CIDR"},
	} {
		t.Run(name, func(t *testing.T) {
			err := New().Validate(map[string]any{
				"features": map[string]any{"bastion": true},
				"bastion":  tc.bastion,
			})
			if tc.want == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want %q", err, tc.want)
			}
		})
	}
}

func TestValidateRequiresVPCDNSForEKS(t *testing.T) {
	cfg := map[string]any{
		"features": map[string]any{"kubernetes_cluster": true},
		"network": map[string]any{
			"availability_zones":   []any{"us-east-1a", "us-east-1b"},
			"enable_dns_hostnames": false,
			"subnets": map[string]any{
				"node": map[string]any{"cidrs": []any{"10.0.1.0/24", "10.0.2.0/24"}, "roles": []any{"cluster"}},
			},
		},
	}
	err := New().Validate(cfg)
	if err == nil || !strings.Contains(err.Error(), "network.enable_dns_hostnames must be true") {
		t.Fatalf("got %v", err)
	}
}
