package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAWSSchemaSectionsHaveNoEnabledProperty(t *testing.T) {
	path := filepath.Join("../../schemas/aws.schema.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(b, &root); err != nil {
		t.Fatal(err)
	}
	props, _ := root["properties"].(map[string]any)
	for _, section := range []string{"kubernetes_cluster", "bastion", "public_dns", "certificates", "network"} {
		sec, ok := props[section].(map[string]any)
		if !ok {
			continue
		}
		inner, _ := sec["properties"].(map[string]any)
		if inner == nil {
			continue
		}
		if _, bad := inner["enabled"]; bad {
			t.Errorf("aws schema section %q must not use enabled (use features.*)", section)
		}
		walkNoEnabled(t, section, inner)
	}
}

func walkNoEnabled(t *testing.T, path string, node map[string]any) {
	if inner, ok := node["properties"].(map[string]any); ok {
		if _, bad := inner["enabled"]; bad {
			t.Errorf("%s: must not define enabled", path)
		}
		for k, v := range inner {
			if m, ok := v.(map[string]any); ok && strings.Contains(k, "") {
				walkNoEnabled(t, path+"."+k, m)
			}
		}
	}
}