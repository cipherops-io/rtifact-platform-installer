package config_test

import (
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud/registry"
	"github.com/cipherops-io/rtifact-platform-installer/internal/version"
)

func TestRegistryPackagedClouds(t *testing.T) {
	if len(registry.Supported) == 0 {
		t.Fatal("expected supported clouds")
	}
	for _, c := range registry.Supported {
		if !version.CloudPackaged(c) {
			t.Fatalf("cloud %q not packaged", c)
		}
		if _, err := registry.Get(c); err != nil {
			t.Fatalf("Get(%q): %v", c, err)
		}
	}
}
