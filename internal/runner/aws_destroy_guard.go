package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
)

func awsDestroyOrphanGuard(customer *config.CustomerConfig, layer stacks.Layer, merged map[string]any) error {
	if customer.Cloud != "aws" {
		return nil
	}
	if layer != stacks.LayerAppPlatform && layer != stacks.LayerCoreNetwork {
		return nil
	}
	state, _ := merged["state"].(map[string]any)
	if air, _ := state["air_gapped"].(bool); air {
		return nil
	}
	kc, _ := merged["kubernetes_cluster"].(map[string]any)
	clusterName, _ := kc["name"].(string)
	if clusterName == "" {
		clusterName = customer.NamePrefix() + "-eks"
	}
	region := customer.Account.Region
	if region == "" {
		return nil
	}
	if _, err := exec.LookPath("aws"); err != nil {
		return nil
	}
	ctx := context.Background()
	hits, err := awsClusterOrphans(ctx, region, clusterName)
	if err != nil {
		return nil
	}
	if len(hits) == 0 {
		return nil
	}
	return usererr.WithHint(
		"Extra AWS resources are still tagged for this cluster.",
		"Remove or retag them before destroy, then retry.",
		strings.Join(hits, "; "),
	)
}

func awsClusterOrphans(ctx context.Context, region, clusterName string) ([]string, error) {
	tagFilter := fmt.Sprintf("tag:kubernetes.io/cluster/%s", clusterName)
	var hits []string
	cmd := exec.CommandContext(ctx, "aws", "ec2", "describe-instances",
		"--region", region,
		"--filters", fmt.Sprintf("Name=tag-key,Values=kubernetes.io/cluster/%s", clusterName), "Name=instance-state-name,Values=running")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var resp struct {
		Reservations []struct {
			Instances []struct {
				InstanceID string `json:"InstanceId"`
			} `json:"Instances"`
		} `json:"Reservations"`
	}
	if json.Unmarshal(out, &resp) == nil {
		for _, r := range resp.Reservations {
			for _, i := range r.Instances {
				hits = append(hits, "EC2 instance "+i.InstanceID)
			}
		}
	}
	_ = tagFilter
	return hits, nil
}