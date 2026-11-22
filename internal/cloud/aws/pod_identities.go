package aws

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
	"github.com/cipherops-io/rtifact-platform-installer/internal/maputil"
)

// Built-in IAM policy documents under infra/modules/aws/pod-identity/policies/<name>.json.
var podIdentityBuiltinPolicies = []string{"external-dns"}

const podIdentityPoliciesDir = "infra/modules/aws/pod-identity/policies"

// ResolvePodIdentities inlines policy_json_file and named policy into policy_json for OpenTofu.
func ResolvePodIdentities(customerPath string, merged map[string]any) error {
	kc, _ := merged["kubernetes_cluster"].(map[string]any)
	if kc == nil {
		return nil
	}
	raw, ok := kc["pod_identities"]
	if !ok || raw == nil {
		return nil
	}
	entries, err := maputil.SliceOfMaps(raw)
	if err != nil {
		return fmt.Errorf("kubernetes_cluster.pod_identities: %w", err)
	}
	if len(entries) == 0 {
		return nil
	}
	contentRoot, _, err := assets.ContentRoot()
	if err != nil {
		return err
	}
	baseDir := filepath.Dir(customerPath)
	resolved := make([]map[string]any, 0, len(entries))
	for i, e := range entries {
		out := maputil.DeepCopyMap(e)
		policyName := strings.TrimSpace(maputil.StringValue(e["policy"]))
		fileRel := strings.TrimSpace(maputil.StringValue(e["policy_json_file"]))
		switch {
		case policyName != "":
			doc, err := readBuiltinPodPolicy(contentRoot, policyName)
			if err != nil {
				return fmt.Errorf("kubernetes_cluster.pod_identities[%d]: %w", i, err)
			}
			out["policy_json"] = doc
			delete(out, "policy")
		case fileRel != "":
			doc, err := readPodPolicyFile(baseDir, fileRel)
			if err != nil {
				return fmt.Errorf("kubernetes_cluster.pod_identities[%d].policy_json_file: %w", i, err)
			}
			out["policy_json"] = doc
		}
		delete(out, "policy_json_file")
		resolved = append(resolved, out)
	}
	outList := make([]any, len(resolved))
	for i, m := range resolved {
		outList[i] = m
	}
	kc["pod_identities"] = outList
	merged["kubernetes_cluster"] = kc
	return nil
}

func readBuiltinPodPolicy(contentRoot, name string) (string, error) {
	if !isBuiltinPodPolicy(name) {
		return "", fmt.Errorf("unknown policy %q (built-in: %s)", name, strings.Join(podIdentityBuiltinPolicies, ", "))
	}
	path := filepath.Join(contentRoot, podIdentityPoliciesDir, name+".json")
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return normalizePolicyJSON(string(b))
}

func readPodPolicyFile(baseDir, rel string) (string, error) {
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("must be relative to the config file directory")
	}
	if strings.Contains(rel, "..") {
		return "", fmt.Errorf("path must not contain ..")
	}
	path := filepath.Join(baseDir, rel)
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return normalizePolicyJSON(string(b))
}

func normalizePolicyJSON(s string) (string, error) {
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	out, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func isBuiltinPodPolicy(name string) bool {
	for _, p := range podIdentityBuiltinPolicies {
		if p == name {
			return true
		}
	}
	return false
}

func validatePodIdentities(features, kc map[string]any) error {
	if !maputil.Truthy(features["pod_identities"]) {
		return nil
	}
	if kc == nil {
		return fmt.Errorf("features.pod_identities is true: set kubernetes_cluster.pod_identities with at least one entry")
	}
	raw, ok := kc["pod_identities"]
	if !ok || raw == nil {
		return fmt.Errorf("features.pod_identities is true: set kubernetes_cluster.pod_identities with at least one entry")
	}
	entries, err := maputil.SliceOfMaps(raw)
	if err != nil {
		return fmt.Errorf("kubernetes_cluster.pod_identities: %w", err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("features.pod_identities is true: kubernetes_cluster.pod_identities must not be empty")
	}
	seen := make(map[string]struct{}, len(entries))
	for i, e := range entries {
		name := maputil.StringValue(e["name"])
		if name == "" {
			return fmt.Errorf("kubernetes_cluster.pod_identities[%d]: name is required", i)
		}
		if _, dup := seen[name]; dup {
			return fmt.Errorf("kubernetes_cluster.pod_identities: duplicate name %q", name)
		}
		seen[name] = struct{}{}
	}
	addons, _ := kc["addons"].(map[string]any)
	if addons == nil {
		return fmt.Errorf("features.pod_identities is true: kubernetes_cluster.addons must include eks-pod-identity-agent")
	}
	if _, ok := addons["eks-pod-identity-agent"]; !ok {
		return fmt.Errorf("features.pod_identities is true: kubernetes_cluster.addons must include eks-pod-identity-agent")
	}
	return nil
}