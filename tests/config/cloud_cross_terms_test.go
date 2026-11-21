package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Azure-only identifiers that must not appear in AWS-owned paths (C6).
var azureTerms = []string{"azurerm", "azure_", "aks_name", "azure_bastion"}
var awsTerms = []string{"aws_eks_cluster", "aws_route53_zone"}

func TestCloudPathsAvoidCrossCloudTerms(t *testing.T) {
	profile := filepath.Join("../../profiles/aios-infra")
	infra := filepath.Join("../../infra")
	checks := []struct {
		root  string
		cloud string
		bad   []string
	}{
		{filepath.Join(profile, "aws"), "aws", azureTerms},
		{filepath.Join(infra, "modules/aws"), "aws", azureTerms},
		{filepath.Join(infra, "stacks/aios-infra/aws"), "aws", azureTerms},
		{filepath.Join(profile, "azure"), "azure", awsTerms},
		{filepath.Join(infra, "modules/azure"), "azure", awsTerms},
		{filepath.Join(infra, "stacks/aios-infra/azure"), "azure", awsTerms},
	}
	for _, c := range checks {
		_ = filepath.Walk(c.root, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".tofu") && !strings.HasSuffix(path, ".md") {
				return nil
			}
			if strings.Contains(path, string(filepath.Separator)+".terraform"+string(filepath.Separator)) {
				return nil
			}
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			text := strings.ToLower(string(b))
			for _, term := range c.bad {
				if strings.Contains(text, strings.ToLower(term)) {
					if allowedCrossCloudTerm(path, term) {
						continue
					}
					t.Errorf("%s (%s path) contains cross-cloud term %q", path, c.cloud, term)
				}
			}
			return nil
		})
	}
}

func allowedCrossCloudTerm(path, term string) bool {
	// Shared docs mention both clouds.
	if strings.HasSuffix(path, "README.md") {
		return true
	}
	// Output contract tests may reference legacy names in comments.
	if strings.Contains(path, "layer_contracts") {
		return true
	}
	_ = term
	return false
}