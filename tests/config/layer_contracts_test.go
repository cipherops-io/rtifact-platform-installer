package config_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

var outputLine = regexp.MustCompile(`^output\s+"([a-z0-9_]+)"`)

func outputsInFile(path string) (map[string]bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	found := map[string]bool{}
	for _, line := range strings.Split(string(b), "\n") {
		if m := outputLine.FindStringSubmatch(strings.TrimSpace(line)); len(m) == 2 {
			found[m[1]] = true
		}
	}
	return found, nil
}

func loadCloudContracts(t *testing.T, cloud string) map[string][]string {
	t.Helper()
	path := filepath.Join("contracts", cloud+".yaml")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read contract %s: %v", path, err)
	}
	out := map[string][]string{}
	if err := yaml.Unmarshal(b, &out); err != nil {
		t.Fatalf("parse contract %s: %v", path, err)
	}
	return out
}

func TestLayerContractsPerCloud(t *testing.T) {
	profile := "aios-infra"
	clouds := []string{"azure", "aws", "gcp"}
	for _, cloud := range clouds {
		byLayer := loadCloudContracts(t, cloud)
		for layer, want := range byLayer {
			path := filepath.Join("../../infra/stacks", profile, cloud, layer, "outputs.tofu")
			if _, err := os.Stat(path); err != nil {
				t.Fatalf("missing outputs for %s/%s: %v", cloud, layer, err)
			}
			got, err := outputsInFile(path)
			if err != nil {
				t.Fatal(err)
			}
			for _, name := range want {
				if !got[name] {
					t.Errorf("%s %s: missing required output %q (see tests/config/contracts/%s.yaml)", cloud, layer, name, cloud)
				}
			}
		}
	}
}
