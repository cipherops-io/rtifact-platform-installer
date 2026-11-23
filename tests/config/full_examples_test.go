package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// Schema keys that examples/<cloud>/full.yaml does not set, with the reason.
var fullExampleSkips = map[string]map[string]string{
	"azure": {
		"network.vpc_cidr":               "AWS network key; not read by the Azure stacks",
		"network.secondary_cidrs":        "AWS network key; not read by the Azure stacks",
		"network.availability_zones":     "AWS network key; not read by the Azure stacks",
		"network.nat_gateway":            "AWS network key; not read by the Azure stacks",
		"network.gateway_endpoints":      "AWS network key; not read by the Azure stacks",
		"network.flow_logs":              "AWS network key; not read by the Azure stacks",
		"network.subnet_prefixes":        "not read by the Azure stacks",
		"kubernetes_cluster.master_cidr": "not read by the Azure stacks",
		"state.bucket_name":              "AWS state bucket",
		"state.encryption":               "AWS state bucket",
		"state.versioning":               "AWS state bucket",
		"bastion.managed":                "legacy flag; bastion.access replaces it",
		"state.lock_table":               "AWS state lock table",
		"network.subnets":                "AWS network key; not read by the Azure stacks",
		"azure_bastion":                  "legacy block; bastion replaces it on Azure",
		"certificates":                   "not read by the Azure stacks",
		"private_dns":                    "not read by the Azure stacks (private_dns_zones is a feature flag)",
	},
	"aws": {
		"state.storage_account": "Azure state storage account",
		"backend":               "the s3 backend is generated from state when bootstrap is true",
	},
	"gcp": {
		"state.storage_account": "Azure state storage account",
		"state.lock_table":      "AWS state lock table",
		"state.bucket_name":     "AWS state bucket",
		"state.encryption":      "AWS state bucket",
		"state.versioning":      "AWS state bucket",
		"state.bootstrap":       "GCP has no state bootstrap",
		"backend":               "local state in the example",
	},
}

// Top-level schema keys that are free-form or shared escape hatches.
var fullExampleOpenSkips = map[string]bool{"overrides": true}

func TestFullExamplesCoverSchema(t *testing.T) {
	root := filepath.Join("..", "..")
	common := loadSchema(t, filepath.Join(root, "schemas", "common.schema.json"))
	for cloud, skips := range fullExampleSkips {
		t.Run(cloud, func(t *testing.T) {
			var paths []string
			walkSchema(common, "", &paths)
			walkSchema(loadSchema(t, filepath.Join(root, "schemas", cloud+".schema.json")), "", &paths)

			raw, err := os.ReadFile(filepath.Join(root, "examples", cloud, "full.yaml"))
			if err != nil {
				t.Fatal(err)
			}
			var doc any
			if err := yaml.Unmarshal(raw, &doc); err != nil {
				t.Fatal(err)
			}

			used := map[string]bool{}
			var missing []string
			seen := map[string]bool{}
			for _, p := range paths {
				if seen[p] {
					continue
				}
				seen[p] = true
				if fullExampleOpenSkips[strings.Split(p, ".")[0]] {
					continue
				}
				if skipFor(skips, p, used) {
					continue
				}
				if !present(doc, strings.Split(p, ".")) {
					missing = append(missing, p)
				}
			}
			sort.Strings(missing)
			if len(missing) > 0 {
				t.Errorf("examples/%s/full.yaml is missing schema keys (add them, or skip them with a reason in full_examples_test.go):\n  %s", cloud, strings.Join(missing, "\n  "))
			}
			for k := range skips {
				if !used[k] {
					t.Errorf("stale skip %q: no longer in the %s schema", k, cloud)
				}
			}
		})
	}
}

func skipFor(skips map[string]string, p string, used map[string]bool) bool {
	for k := range skips {
		if p == k || strings.HasPrefix(p, k+".") {
			used[k] = true
			return true
		}
	}
	return false
}

func loadSchema(t *testing.T, path string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var s map[string]any
	if err := json.Unmarshal(b, &s); err != nil {
		t.Fatal(err)
	}
	return s
}

// walkSchema records the leaf paths of a schema. "*" stands for a map key or an array item;
// objects without declared properties (open blocks) are leaves.
func walkSchema(s map[string]any, path string, out *[]string) {
	join := func(k string) string {
		if path == "" {
			return k
		}
		return path + "." + k
	}
	props, _ := s["properties"].(map[string]any)
	ap, apIsSchema := s["additionalProperties"].(map[string]any)
	switch {
	case len(props) > 0 || apIsSchema:
		for k, v := range props {
			if vs, ok := v.(map[string]any); ok {
				walkSchema(vs, join(k), out)
			}
		}
		if apIsSchema {
			walkSchema(ap, join("*"), out)
		}
	case s["type"] == "array":
		if items, ok := s["items"].(map[string]any); ok {
			if _, hasProps := items["properties"]; hasProps {
				walkSchema(items, join("*"), out)
				return
			}
		}
		*out = append(*out, path)
	default:
		if path != "" {
			*out = append(*out, path)
		}
	}
}

// present reports whether the path exists in the YAML; "*" matches any map entry or array item.
func present(node any, parts []string) bool {
	if len(parts) == 0 {
		return true
	}
	if parts[0] == "*" {
		switch n := node.(type) {
		case map[string]any:
			for _, v := range n {
				if present(v, parts[1:]) {
					return true
				}
			}
		case []any:
			for _, v := range n {
				if present(v, parts[1:]) {
					return true
				}
			}
		}
		return false
	}
	m, ok := node.(map[string]any)
	if !ok {
		return false
	}
	v, ok := m[parts[0]]
	if !ok {
		return false
	}
	return present(v, parts[1:])
}
