package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergeBuildResolvesBuiltinPodPolicy(t *testing.T) {
	root := repoRoot(t)
	t.Setenv("INSTALLER_CONTENT_ROOT", root)
	cfg := filepath.Join(root, "examples", "aws", "aws-sbx2.yaml")
	profilesDir := filepath.Join(root, "profiles")
	defaults, err := LoadProfileDefaults(profilesDir, "aios-infra", "aws")
	if err != nil {
		t.Fatal(err)
	}
	merged, err := MergeBuildFromPath(defaults, cfg)
	if err != nil {
		t.Fatal(err)
	}
	kc, _ := merged["kubernetes_cluster"].(map[string]any)
	entries, _ := kc["pod_identities"].([]any)
	if len(entries) != 1 {
		t.Fatalf("pod_identities len = %d", len(entries))
	}
	e, _ := entries[0].(map[string]any)
	if e["policy"] != nil {
		t.Fatalf("policy key should be stripped after resolve, got %#v", e["policy"])
	}
	pj, _ := e["policy_json"].(string)
	if pj == "" {
		t.Fatal("expected policy_json after resolving builtin policy")
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(pj), &doc); err != nil {
		t.Fatal(err)
	}
}

func TestMergeBuildResolvesPodPolicyJSONFile(t *testing.T) {
	root := repoRoot(t)
	t.Setenv("INSTALLER_CONTENT_ROOT", root)
	dir := t.TempDir()
	policyPath := filepath.Join(dir, "extra-policy.json")
	if err := os.WriteFile(policyPath, []byte(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:ListAllMyBuckets","Resource":"*"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(dir, "cfg.yaml")
	cfgYAML := `schema_version: 1
cloud: aws
profile: aios-infra
tier: sandbox
customer: { name: T, code: t }
account: { account_id: "123456789012", region: us-east-1 }
workload: { name: w, environment: e, name_prefix: pfx }
state: { backend: s3, bootstrap: true, air_gapped: true }
features:
  network: true
  kubernetes_cluster: true
  pod_identities: true
network:
  vpc_cidr: 10.0.0.0/16
  availability_zones: [us-east-1a, us-east-1b]
  subnets:
    public:
      cidrs: [10.0.0.0/24, 10.0.1.0/24]
      route: internet
    private:
      cidrs: [10.0.2.0/24, 10.0.3.0/24]
      route: nat
    node:
      cidrs: [10.0.4.0/24, 10.0.5.0/24]
      route: nat
      roles: [cluster, nodes]
kubernetes_cluster:
  endpoint: { public: false, private: true }
  addons:
    eks-pod-identity-agent: { version: v1 }
  pod_identities:
    - name: demo
      namespace: demo
      service_account: demo
      policy_json_file: extra-policy.json
`
	if err := os.WriteFile(cfgPath, []byte(cfgYAML), 0o600); err != nil {
		t.Fatal(err)
	}
	profilesDir := filepath.Join(root, "profiles")
	defaults, err := LoadProfileDefaults(profilesDir, "aios-infra", "aws")
	if err != nil {
		t.Fatal(err)
	}
	merged, err := MergeBuildFromPath(defaults, cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	kc, _ := merged["kubernetes_cluster"].(map[string]any)
	entries, _ := kc["pod_identities"].([]any)
	e, _ := entries[0].(map[string]any)
	if !strings.Contains(e["policy_json"].(string), "s3:ListAllMyBuckets") {
		t.Fatalf("unexpected policy_json: %s", e["policy_json"])
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}