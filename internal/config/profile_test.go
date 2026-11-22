package config

import (
	"path/filepath"
	"sync"
	"testing"
)

func TestResolveProfileID(t *testing.T) {
	root := filepath.Join("..", "..")
	t.Setenv("INSTALLER_CONTENT_ROOT", root)
	aliasOnce = sync.Once{}
	aliasIndex = nil
	aliasErr = nil

	if got := ResolveProfileID("ai-os-k8s"); got != "aios-infra" {
		t.Fatalf("got %q", got)
	}
	if got := ResolveProfileID("aios-infra"); got != "aios-infra" {
		t.Fatalf("got %q", got)
	}
}
