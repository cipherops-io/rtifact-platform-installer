package config_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var outputLine = regexp.MustCompile(`^output\s+"([a-z0-9_]+)"`)

// allCloudOutputs must appear in every cloud's outputs.tofu for the given layer.
var allCloudOutputs = map[string][]string{
	"state-backend": {
		"resource_group_name",
	},
	"core-network": {
		"resource_group_name", "resource_group_id", "tags", "vnet_id",
		"node_subnet_id", "load_balancer_subnet_id", "private_endpoint_subnet_id", "management_subnet_id",
	},
	"data-platform": {
		"key_vault_id", "key_vault_uri", "postgresql_id", "storage_account_id", "acr_id",
	},
	"app-platform": {
		"cluster_name", "aks_name", "oidc_issuer_url", "bastion_private_ip", "bastion_ssh_private_key_pem",
	},
}

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

func TestLayerContractsAllCloudOutputs(t *testing.T) {
	profile := "aios-infra"
	clouds := []string{"azure", "aws", "gcp"}
	for layer, want := range allCloudOutputs {
		for _, cloud := range clouds {
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
					t.Errorf("%s %s: missing required output %q (enforced by docs/layer-contracts.md)", cloud, layer, name)
				}
			}
		}
	}
}