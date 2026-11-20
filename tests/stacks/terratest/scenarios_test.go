package terratest

import (
	"fmt"
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"

	"github.com/cipherops-io/boilerplates/devops-templates/tofu-stack/terratest/testutil"
)

var layersNeedUpstream = map[string]bool{
	"data-platform":   true,
	"app-platform":    true,
	"network-edge":    true,
	"ai-integrations": true,
}

func init() {
	clouds := []string{"azure", "aws", "gcp"}
	planLayers := []string{
		"state-backend",
		"core-network",
		"data-platform",
		"app-platform",
		"network-edge",
		"ai-integrations",
	}

	for _, cloud := range clouds {
		varFile := testutil.VarFileForCloud(cloud)
		for _, layer := range planLayers {
			cloudName, layerName := cloud, layer
			testCases = append(testCases, TestCase{
				Name:     fmt.Sprintf("validate/%s/%s", cloudName, layerName),
				Parallel: true,
				Options: func(t *testing.T) *terraform.Options {
					dir, fixture := testutil.SandboxLayer(t, cloudName, layerName)
					opts := defaultPlanOptions(t, cloudName, dir, varFile)
					if layersNeedUpstream[layerName] {
						opts.Vars = map[string]interface{}{
							"upstream_state": map[string]interface{}{
								"core-network": map[string]interface{}{
									"backend": "local",
									"config": map[string]string{
										"path": fixture,
									},
								},
							},
						}
					}
					return opts
				},
				Run: initAndValidateRun,
			})
		}

		if os.Getenv("TERRATEST_PLAN") == "true" {
			for _, layer := range planLayers {
				cloudName, layerName := cloud, layer
				testCases = append(testCases, TestCase{
						Name:     fmt.Sprintf("plan/%s/%s", cloudName, layerName),
						Parallel: true,
						Options: func(t *testing.T) *terraform.Options {
							dir, fixture := testutil.SandboxLayer(t, cloudName, layerName)
							opts := defaultPlanOptions(t, cloudName, dir, varFile)
							if layersNeedUpstream[layerName] {
								opts.Vars = map[string]interface{}{
									"upstream_state": map[string]interface{}{
										"core-network": map[string]interface{}{
											"backend": "local",
											"config": map[string]string{
												"path": fixture,
											},
										},
									},
								}
							}
							return opts
						},
						Run: initAndPlanRun,
					})
			}
		}

		testCases = append(testCases, TestCase{
			Name: fmt.Sprintf("e2e/%s/state-backend", cloud),
			Options: func(t *testing.T) *terraform.Options {
				dir, _ := testutil.SandboxLayer(t, cloud, "state-backend")
				return defaultPlanOptions(t, cloud, dir, varFile)
			},
			Run: initAndApplyAndDestroyRun,
		})
	}
}