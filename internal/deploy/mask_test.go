package deploy

import (
	"strings"
	"testing"
)

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

func TestResolvedSecretRedactor(t *testing.T) {
	var r ResolvedSecretRedactor
	r.RecordResolved("s3cret")
	got := r.RedactText("failed: s3cret leaked and ${env.X}")
	if strings.Contains(got, "s3cret") || strings.Contains(got, "${env.") {
		t.Fatalf("redact: %q", got)
	}
}