package deploy

import "testing"

func TestMaskEnvPlaceholders(t *testing.T) {
	in := "password: ${env.PLATFORM_PG_PASSWORD}"
	got := MaskEnvPlaceholders(in)
	if got == in || got == "" {
		t.Fatalf("expected masking, got %q", got)
	}
	if got != "password: ***" {
		t.Fatalf("got %q", got)
	}
}
