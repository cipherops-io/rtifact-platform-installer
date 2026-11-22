package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPodIdentityBuiltinPoliciesExist(t *testing.T) {
	root := filepath.Join("..", "..")
	schemaPath := filepath.Join(root, "schemas", "aws.schema.json")
	b, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatal(err)
	}
	var sch map[string]any
	if err := json.Unmarshal(b, &sch); err != nil {
		t.Fatal(err)
	}
	kc := sch["properties"].(map[string]any)["kubernetes_cluster"].(map[string]any)
	items := kc["properties"].(map[string]any)["pod_identities"].(map[string]any)["items"].(map[string]any)
	policy := items["properties"].(map[string]any)["policy"].(map[string]any)
	enumRaw, _ := policy["enum"].([]any)
	var names []string
	for _, v := range enumRaw {
		names = append(names, v.(string))
	}
	policiesDir := filepath.Join(root, "infra", "modules", "aws", "pod-identity", "policies")
	for _, name := range names {
		path := filepath.Join(policiesDir, name+".json")
		doc, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("missing built-in policy %s: %v", path, err)
		}
		var v any
		if err := json.Unmarshal(doc, &v); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !strings.Contains(string(doc), "Statement") {
			t.Fatalf("%s: expected IAM policy document", name)
		}
	}
}