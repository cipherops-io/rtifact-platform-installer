package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/maputil"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
)

func (p *Provider) DestroyOrphanGuard(layer stacks.Layer, region, namePrefix string, merged map[string]any) error {
	if layer != stacks.LayerAppPlatform && layer != stacks.LayerCoreNetwork {
		return nil
	}
	state, _ := merged["state"].(map[string]any)
	if air, _ := state["air_gapped"].(bool); air {
		return nil
	}
	kc, _ := merged["kubernetes_cluster"].(map[string]any)
	clusterName := maputil.StringValue(kc["name"])
	if clusterName == "" {
		clusterName = namePrefix + "-eks"
	}
	if region == "" {
		return nil
	}
	if _, err := p.lookPath(CLI); err != nil {
		return nil
	}
	ctx := context.Background()
	hits, err := awsClusterOrphans(ctx, p, region, clusterName)
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

func (p *Provider) MutateDNSDelegation(ctx context.Context, layer stacks.Layer, region, namePrefix, networkEdgeOutputsPath string, merged map[string]any) {
	if layer != stacks.LayerNetworkEdge {
		return
	}
	if dnsDelegationVerified(ctx, p, region, namePrefix, networkEdgeOutputsPath, merged) {
		merged["dns_delegation_verified"] = true
	}
}

func (p *Provider) DelegationPauseIfNeeded(merged map[string]any, configPath, _ string, flat map[string]any) error {
	features, _ := merged["features"].(map[string]any)
	if !maputil.Truthy(features["public_dns"]) || !maputil.Truthy(features["certificates"]) {
		return nil
	}
	pdns, _ := merged["public_dns"].(map[string]any)
	if maputil.StringValue(pdns["parent_zone_id"]) != "" || maputil.Truthy(pdns["delegated"]) {
		return nil
	}
	ns := maputil.FlattenStringSlice(flat["dns_name_servers"])
	if len(ns) == 0 {
		return nil
	}
	if arn := maputil.StringValue(flat["certificate_regional_arn"]); arn != "" {
		return nil
	}
	zone := maputil.StringValue(pdns["zone_name"])
	var b strings.Builder
	fmt.Fprintf(&b, "Created Route53 zone for %q. Add these NS records at the parent DNS:\n\n", zone)
	for _, n := range ns {
		fmt.Fprintf(&b, "  %s\n", n)
	}
	fmt.Fprintf(&b, "\nThen set public_dns.delegated: true and run:\n\n  rtifact apply %s --stack network-edge --yes\n", configPath)
	return &usererr.Error{
		Summary: "DNS delegation required before ACM certificates.",
		Hint:    b.String(),
	}
}

func dnsDelegationVerified(ctx context.Context, p *Provider, region, namePrefix, outputsPath string, merged map[string]any) bool {
	pdns, _ := merged["public_dns"].(map[string]any)
	if pdns == nil {
		return false
	}
	if maputil.StringValue(pdns["parent_zone_id"]) != "" || maputil.Truthy(pdns["delegated"]) {
		return true
	}
	zone := maputil.StringValue(pdns["zone_name"])
	if zone == "" {
		return false
	}
	expected := readSavedNameServers(outputsPath)
	if len(expected) == 0 {
		expected = route53ZoneNameServers(ctx, p, region, zone)
	}
	if len(expected) == 0 {
		return false
	}
	live, err := net.DefaultResolver.LookupNS(ctx, zone)
	if err != nil || len(live) == 0 {
		return false
	}
	liveSet := nsHostSet(live)
	for _, e := range expected {
		h := strings.ToLower(strings.TrimSuffix(e, "."))
		if !liveSet[h] {
			return false
		}
	}
	return true
}

func route53ZoneNameServers(ctx context.Context, p *Provider, region, zone string) []string {
	if _, err := p.lookPath(CLI); err != nil {
		return nil
	}
	if region == "" {
		return nil
	}
	listOut, err := p.run(ctx, CLI, "route53", "list-hosted-zones-by-name",
		"--dns-name", zone,
		"--max-items", "1",
		"--output", "json",
	)
	if err != nil {
		return nil
	}
	var listed struct {
		HostedZones []struct {
			ID   string `json:"Id"`
			Name string `json:"Name"`
		} `json:"HostedZones"`
	}
	if json.Unmarshal(listOut, &listed) != nil || len(listed.HostedZones) == 0 {
		return nil
	}
	zid := strings.TrimPrefix(listed.HostedZones[0].ID, "/hostedzone/")
	hzOut, err := p.run(ctx, CLI, "route53", "get-hosted-zone",
		"--id", zid,
		"--output", "json",
	)
	if err != nil {
		return nil
	}
	var hz struct {
		DelegationSet struct {
			NameServers []string `json:"NameServers"`
		} `json:"DelegationSet"`
	}
	if json.Unmarshal(hzOut, &hz) != nil {
		return nil
	}
	return hz.DelegationSet.NameServers
}

func readSavedNameServers(outputsPath string) []string {
	b, err := os.ReadFile(outputsPath)
	if err != nil {
		return nil
	}
	var flat map[string]any
	if json.Unmarshal(b, &flat) != nil {
		return nil
	}
	return maputil.FlattenStringSlice(flat["dns_name_servers"])
}

func nsHostSet(records []*net.NS) map[string]bool {
	out := map[string]bool{}
	for _, r := range records {
		h := strings.ToLower(strings.TrimSuffix(r.Host, "."))
		out[h] = true
	}
	return out
}

func awsClusterOrphans(ctx context.Context, p *Provider, region, clusterName string) ([]string, error) {
	tagKey := fmt.Sprintf("kubernetes.io/cluster/%s", clusterName)
	var hits []string

	instOut, err := p.run(ctx, CLI, "ec2", "describe-instances",
		"--region", region,
		"--filters",
		"Name=tag-key,Values="+tagKey,
		"Name=instance-state-name,Values=running,stopped",
	)
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

	lbOut, err := p.run(ctx, CLI, "elbv2", "describe-load-balancers", "--region", region)
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
			tagsOut, err := p.run(ctx, CLI, "elbv2", "describe-tags",
				"--region", region,
				"--resource-arns", lb.ARN,
			)
			if err != nil {
				continue
			}
			if strings.Contains(string(tagsOut), tagKey) || strings.Contains(string(tagsOut), clusterName) {
				hits = append(hits, "Load balancer "+lb.Name)
			}
		}
	}

	volOut, err := p.run(ctx, CLI, "ec2", "describe-volumes",
		"--region", region,
		"--filters", "Name=tag-key,Values="+tagKey,
	)
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