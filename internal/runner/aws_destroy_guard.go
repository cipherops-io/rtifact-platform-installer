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
		return fmt.Errorf("aws destroy guard: %w", err)
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
	tagKey := fmt.Sprintf("kubernetes.io/cluster/%s", clusterName)
	var hits []string

	instOut, err := exec.CommandContext(ctx, "aws", "ec2", "describe-instances",
		"--region", region,
		"--filters",
		"Name=tag-key,Values="+tagKey,
		"Name=instance-state-name,Values=running,stopped",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("describe-instances: %w", err)
	}
	var instResp struct {
		Reservations []struct {
			Instances []struct {
				InstanceID string `json:"InstanceId"`
			} `json:"Instances"`
		} `json:"Reservations"`
	}
	if json.Unmarshal(instOut, &instResp) == nil {
		for _, r := range instResp.Reservations {
			for _, i := range r.Instances {
				hits = append(hits, "EC2 instance "+i.InstanceID)
			}
		}
	}

	lbOut, err := exec.CommandContext(ctx, "aws", "elbv2", "describe-load-balancers", "--region", region).Output()
	if err != nil {
		return nil, fmt.Errorf("describe-load-balancers: %w", err)
	}
	var lbResp struct {
		LoadBalancers []struct {
			Name string `json:"LoadBalancerName"`
			ARN  string `json:"LoadBalancerArn"`
		} `json:"LoadBalancers"`
	}
	if json.Unmarshal(lbOut, &lbResp) == nil {
		for _, lb := range lbResp.LoadBalancers {
			tagsOut, err := exec.CommandContext(ctx, "aws", "elbv2", "describe-tags",
				"--region", region,
				"--resource-arns", lb.ARN,
			).Output()
			if err != nil {
				continue
			}
			if strings.Contains(string(tagsOut), tagKey) || strings.Contains(string(tagsOut), clusterName) {
				hits = append(hits, "Load balancer "+lb.Name)
			}
		}
	}

	volOut, err := exec.CommandContext(ctx, "aws", "ec2", "describe-volumes",
		"--region", region,
		"--filters", "Name=tag-key,Values="+tagKey,
	).Output()
	if err != nil {
		return nil, fmt.Errorf("describe-volumes: %w", err)
	}
	var volResp struct {
		Volumes []struct {
			VolumeID string `json:"VolumeId"`
		} `json:"Volumes"`
	}
	if json.Unmarshal(volOut, &volResp) == nil {
		for _, v := range volResp.Volumes {
			hits = append(hits, "EBS volume "+v.VolumeID)
		}
	}

	return hits, nil
}