package aws

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/maputil"
)

func (*Provider) ExpandCustomerPaths(customerPath string, merged map[string]any) error {
	return ResolvePodIdentities(customerPath, merged)
}

func (*Provider) TFVarSections() []string {
	return []string{
		"features", "state", "tags", "certificates", "network",
		"kubernetes_cluster", "bastion", "public_dns",
	}
}

func (*Provider) Normalize(merged map[string]any) {
	bastion, _ := merged["bastion"].(map[string]any)
	if bastion == nil {
		bastion = make(map[string]any)
	}
	if _, ok := bastion["access"]; !ok {
		bastion["access"] = "ssm"
	}
	merged["bastion"] = bastion

	kc, _ := merged["kubernetes_cluster"].(map[string]any)
	if kc != nil {
		if se, ok := kc["secrets_encryption"].(map[string]any); ok {
			if _, hasMode := se["mode"]; !hasMode {
				if maputil.Truthy(se["enabled"]) {
					se["mode"] = "kms"
				} else {
					se["mode"] = "none"
				}
				kc["secrets_encryption"] = se
			}
		}
		merged["kubernetes_cluster"] = kc
	}
}

func (*Provider) Validate(merged map[string]any) error {
	features, _ := merged["features"].(map[string]any)
	if err := validateAWSSubnets(merged); err != nil {
		return err
	}
	if maputil.Truthy(features["network"]) {
		netw, ok := merged["network"].(map[string]any)
		if !ok {
			return fmt.Errorf("features.network is true: set a network block with vpc_cidr, two availability_zones, and subnets")
		}
		if _, ok := netw["vpc_cidr"].(string); !ok || netw["vpc_cidr"] == "" {
			return fmt.Errorf("features.network: network.vpc_cidr is required")
		}
		zones, _ := maputil.StringSlice(netw["availability_zones"])
		if len(zones) < 2 {
			return fmt.Errorf("features.network: network.availability_zones must list at least two zones")
		}
		subnets, ok := netw["subnets"].(map[string]any)
		if !ok || len(subnets) == 0 {
			return fmt.Errorf("features.network: network.subnets is required")
		}
	}
	var kc map[string]any
	if maputil.Truthy(features["kubernetes_cluster"]) {
		netw, _ := merged["network"].(map[string]any)
		subnets, _ := netw["subnets"].(map[string]any)
		zones, _ := maputil.StringSlice(netw["availability_zones"])
		if !subnetTierCoversZones(subnets, zones, "cluster") {
			return fmt.Errorf("features.kubernetes_cluster: include a subnet tier with role cluster in two zones")
		}
		var ok bool
		kc, ok = merged["kubernetes_cluster"].(map[string]any)
		if !ok {
			return fmt.Errorf("features.kubernetes_cluster is true: set kubernetes_cluster")
		}
		if ep, ok := kc["endpoint"].(map[string]any); ok && maputil.Truthy(ep["public"]) {
			cidrs, _ := maputil.StringSlice(ep["public_access_cidrs"])
			if len(cidrs) == 0 {
				return fmt.Errorf("kubernetes_cluster.endpoint.public is true: set endpoint.public_access_cidrs")
			}
		}
		if se, ok := kc["secrets_encryption"].(map[string]any); ok {
			mode, _ := se["mode"].(string)
			if mode == "kms" && strings.TrimSpace(fmt.Sprint(se["kms_key_arn"])) == "" {
				return fmt.Errorf("kubernetes_cluster.secrets_encryption.mode kms requires kms_key_arn")
			}
		}
	}
	if kc == nil {
		kc, _ = merged["kubernetes_cluster"].(map[string]any)
	}
	if err := validatePodIdentities(features, kc); err != nil {
		return err
	}
	return nil
}

func (p *Provider) Preflight(ctx context.Context, merged map[string]any) error {
	state, _ := merged["state"].(map[string]any)
	if airGapped, _ := state["air_gapped"].(bool); airGapped {
		return nil
	}
	account, _ := merged["account"].(map[string]any)
	region, _ := account["region"].(string)
	if region == "" {
		return fmt.Errorf("preflight: account.region is required")
	}
	if _, err := p.lookPath(CLI); err != nil {
		return fmt.Errorf("preflight: aws CLI not found in PATH (required when state.air_gapped is false)")
	}
	cmd := exec.CommandContext(ctx, "aws", "sts", "get-caller-identity", "--region", region)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("preflight: AWS credentials: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	netw, _ := merged["network"].(map[string]any)
	var cidrs []string
	if vpc, _ := netw["vpc_cidr"].(string); vpc != "" {
		cidrs = append(cidrs, vpc)
	}
	if secs, err := maputil.StringSlice(netw["secondary_cidrs"]); err == nil {
		cidrs = append(cidrs, secs...)
	}
	for _, cidr := range cidrs {
		if err := awsVPCCIDRFree(ctx, region, cidr); err != nil {
			return err
		}
	}
	return nil
}

func awsVPCCIDRFree(ctx context.Context, region, vpcCIDR string) error {
	cmd := exec.CommandContext(ctx, "aws", "ec2", "describe-vpcs", "--region", region, "--filters", "Name=cidr-block-association-set.cidr-block,Values="+vpcCIDR)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("preflight: describe-vpcs for %s: %w (%s)", vpcCIDR, err, strings.TrimSpace(string(out)))
	}
	if strings.Contains(string(out), vpcCIDR) {
		return fmt.Errorf("preflight: an existing VPC already uses CIDR %s in %s", vpcCIDR, region)
	}
	return nil
}

func subnetTierCoversZones(subnets map[string]any, zones []string, wantRole string) bool {
	for tier, raw := range subnets {
		tierMap, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		nameMatch := tier == wantRole
		roles, _ := maputil.StringSlice(tierMap["roles"])
		for _, r := range roles {
			if r == wantRole {
				nameMatch = true
			}
		}
		if !nameMatch {
			continue
		}
		cidrs, err := maputil.StringSlice(tierMap["cidrs"])
		if err != nil || len(cidrs) < 2 {
			continue
		}
		if len(zones) >= 2 && len(cidrs) >= 2 {
			return true
		}
	}
	return false
}

func validateAWSSubnets(merged map[string]any) error {
	netw, ok := merged["network"].(map[string]any)
	if !ok {
		return nil
	}
	zones, _ := maputil.StringSlice(netw["availability_zones"])
	if len(zones) > 0 && len(zones) < 2 {
		return fmt.Errorf("network.availability_zones must list at least two zones for AWS")
	}
	subnets, ok := netw["subnets"].(map[string]any)
	if !ok {
		return nil
	}
	vpcCIDR, _ := netw["vpc_cidr"].(string)
	secondaries, _ := maputil.StringSlice(netw["secondary_cidrs"])
	var allowed []*net.IPNet
	if vpcCIDR != "" {
		_, n, err := net.ParseCIDR(vpcCIDR)
		if err != nil {
			return fmt.Errorf("network.vpc_cidr: %w", err)
		}
		allowed = append(allowed, n)
	}
	for _, c := range secondaries {
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			return fmt.Errorf("network.secondary_cidrs: invalid %s", c)
		}
		allowed = append(allowed, n)
	}
	var allTierCIDRs []string
	for tier, raw := range subnets {
		tierMap, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		cidrs, err := maputil.StringSlice(tierMap["cidrs"])
		if err != nil {
			return fmt.Errorf("network.subnets.%s.cidrs: %w", tier, err)
		}
		if len(zones) > 0 && len(cidrs) != len(zones) {
			return fmt.Errorf("network.subnets.%s.cidrs: want %d entries, got %d", tier, len(zones), len(cidrs))
		}
		for _, c := range cidrs {
			allTierCIDRs = append(allTierCIDRs, c)
			if len(allowed) > 0 && !cidrInAllowed(c, allowed) {
				return fmt.Errorf("network.subnets.%s: CIDR %s is not inside vpc_cidr or secondary_cidrs", tier, c)
			}
		}
	}
	return checkCIDROverlaps(allTierCIDRs)
}