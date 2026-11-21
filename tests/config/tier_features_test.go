package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestTierFilesDoNotSetFeatures(t *testing.T) {
	root := filepath.Join("../../profiles", "aios-infra")
	clouds := []string{"azure", "aws", "gcp"}
	for _, cloud := range clouds {
		tiersDir := filepath.Join(root, cloud, "tiers")
		entries, err := os.ReadDir(tiersDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			t.Fatalf("%s: %v", tiersDir, err)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
				continue
			}
			path := filepath.Join(tiersDir, e.Name())
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var doc map[string]any
			if err := yaml.Unmarshal(b, &doc); err != nil {
				t.Fatalf("%s: %v", path, err)
			}
			if _, ok := doc["features"]; ok {
				t.Errorf("%s must not contain a features block (enable features in customer config)", path)
			}
		}
	}
}