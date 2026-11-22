package deploy

import "testing"

func TestCheckPlaceholderString(t *testing.T) {
	tests := []struct {
		in    string
		ok    bool
		label string
	}{
		{"${DS_PROMETHEUS}", true, "chart-owned"},
		{"prefix ${DS_X} suffix", true, "chart-owned embedded"},
		{"${env.PLATFORM_PG_PASSWORD}", true, "valid env"},
		{"${env.A}", true, "single-letter env name"},
		{"${env.lower}", false, "lowercase env"},
		{"${infra.public_dns_zone_id}", true, "valid infra"},
		{"${infra.}", false, "empty infra key"},
	}
	for _, tc := range tests {
		err := checkPlaceholderString(tc.in)
		if tc.ok && err != nil {
			t.Fatalf("%s: %q: unexpected error: %v", tc.label, tc.in, err)
		}
		if !tc.ok && err == nil {
			t.Fatalf("%s: %q: expected error", tc.label, tc.in)
		}
	}
}

func TestMaskEnvPlaceholderSingleLetter(t *testing.T) {
	got := MaskEnvPlaceholders("x=${env.A}")
	if got != "x=***" {
		t.Fatalf("got %q", got)
	}
}