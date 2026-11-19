package config

import "testing"

func TestResolveProfileID(t *testing.T) {
	if got := ResolveProfileID("ai-os-k8s"); got != "aios-infra" {
		t.Fatalf("got %q", got)
	}
	if got := ResolveProfileID("aios-infra"); got != "aios-infra" {
		t.Fatalf("got %q", got)
	}
}