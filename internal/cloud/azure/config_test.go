package azure

import (
	"strings"
	"testing"
)

func TestNormalizeBastionAccess(t *testing.T) {
	cases := []struct {
		name    string
		bastion map[string]any
		want    string
	}{
		{"default", nil, "private"},
		{"legacy managed", map[string]any{"managed": true}, "azure_bastion"},
		{"explicit wins", map[string]any{"managed": true, "access": "public_ip"}, "public_ip"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			merged := map[string]any{}
			if tc.bastion != nil {
				merged["bastion"] = tc.bastion
			}
			New().Normalize(merged)
			b := merged["bastion"].(map[string]any)
			if b["access"] != tc.want {
				t.Fatalf("access = %v, want %s", b["access"], tc.want)
			}
		})
	}
}

func TestValidateBastionAccess(t *testing.T) {
	base := func(bastion map[string]any, network map[string]any) map[string]any {
		return map[string]any{
			"features": map[string]any{"azure_bastion": true},
			"bastion":  bastion,
			"network":  network,
		}
	}
	cases := []struct {
		name   string
		merged map[string]any
		errSub string
	}{
		{"private ok", base(map[string]any{"access": "private"}, nil), ""},
		{"public needs cidrs", base(map[string]any{"access": "public_ip"}, nil), "allowed_ssh_cidrs"},
		{"public ok", base(map[string]any{"access": "public_ip", "allowed_ssh_cidrs": []any{"203.0.113.10/32"}}, nil), ""},
		{"azure bastion needs subnet", base(map[string]any{"access": "azure_bastion"}, map[string]any{}), "azure_bastion_subnet_cidr"},
		{"azure bastion subnet too small", base(map[string]any{"access": "azure_bastion"}, map[string]any{"azure_bastion_subnet_cidr": "10.20.8.0/27"}), "/26 or larger"},
		{"azure bastion ok", base(map[string]any{"access": "azure_bastion"}, map[string]any{"azure_bastion_subnet_cidr": "10.20.8.0/26"}), ""},
		{"unknown access", base(map[string]any{"access": "vpn"}, nil), "must be private"},
		{"feature off ignores", map[string]any{"features": map[string]any{"azure_bastion": false}, "bastion": map[string]any{"access": "vpn"}}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := New().Validate(tc.merged)
			if tc.errSub == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.errSub) {
				t.Fatalf("error = %v, want substring %q", err, tc.errSub)
			}
		})
	}
}
