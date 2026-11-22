package config

import (
	"path/filepath"
	"testing"
)

func TestWriteInitConfigValidatesEachCloud(t *testing.T) {
	for _, cloud := range []string{"azure", "aws", "gcp"} {
		t.Run(cloud, func(t *testing.T) {
			dir := t.TempDir()
			out := filepath.Join(dir, "cfg.yaml")
			if err := WriteInitConfig(cloud, "sandbox", "Acme", out); err != nil {
				t.Fatal(err)
			}
			if err := ValidateCustomerYAML(out); err != nil {
				t.Fatalf("init output should validate: %v", err)
			}
		})
	}
}