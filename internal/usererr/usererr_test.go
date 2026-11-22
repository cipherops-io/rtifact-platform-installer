package usererr

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestHumanStripsCobraUsage(t *testing.T) {
	err := fmt.Errorf("quota insufficient\n\nUsage:\n  validate [flags]")
	got := Human(err)
	if strings.Contains(got, "Usage:") {
		t.Fatalf("usage not stripped: %q", got)
	}
}

func TestWrapPreservesUsererr(t *testing.T) {
	inner := WithHint("a", "hint", "detail")
	out := Wrap(inner, "ignored", "")
	var ue *Error
	if !errors.As(out, &ue) || ue.Summary != "a" {
		t.Fatalf("wrap changed user err: %v", out)
	}
}

func TestPrettifyFeaturesUsesProfile(t *testing.T) {
	SetProfileID("my-profile")
	t.Cleanup(func() { SetProfileID("") })
	got := Human(fmt.Errorf("features.foo: unknown feature (not in registry)"))
	if !strings.Contains(got, "profiles/my-profile/features.yaml") {
		t.Fatalf("expected dynamic profile path: %q", got)
	}
}