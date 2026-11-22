package runner

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

func injectDNSDelegation(ctx context.Context, layer stacks.Layer, customer *config.CustomerConfig, merged map[string]any) {
	if layer != stacks.LayerNetworkEdge || customer.Cloud != "aws" {
		return
	}
	if dnsDelegationVerified(ctx, customer, merged) {
		merged["dns_delegation_verified"] = true
	}
}

func dnsDelegationVerified(ctx context.Context, customer *config.CustomerConfig, merged map[string]any) bool {
	pdns, _ := merged["public_dns"].(map[string]any)
	if pdns == nil {
		return false
	}
	if stringValue(pdns["parent_zone_id"]) != "" || truthy(pdns["delegated"]) {
		return true
	}
	zone, _ := pdns["zone_name"].(string)
	if zone == "" {
		return false
	}
	expected := readSavedNameServers(customer.NamePrefix())
	if len(expected) == 0 {
		expected = route53ZoneNameServers(ctx, customer.Account.Region, zone)
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

func route53ZoneNameServers(ctx context.Context, region, zone string) []string {
	if _, err := exec.LookPath("aws"); err != nil {
		return nil
	}
	if region == "" {
		return nil
	}
	listOut, err := exec.CommandContext(ctx, "aws", "route53", "list-hosted-zones-by-name",
		"--dns-name", zone,
		"--max-items", "1",
		"--output", "json",
	).Output()
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
	hzOut, err := exec.CommandContext(ctx, "aws", "route53", "get-hosted-zone",
		"--id", zid,
		"--output", "json",
	).Output()
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

func readSavedNameServers(prefix string) []string {
	path := filepath.Join(config.ArtifactsDir(prefix), string(stacks.LayerNetworkEdge)+"-outputs.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var flat map[string]any
	if json.Unmarshal(b, &flat) != nil {
		return nil
	}
	return flattenStringSlice(flat["dns_name_servers"])
}

func nsHostSet(records []*net.NS) map[string]bool {
	out := map[string]bool{}
	for _, r := range records {
		h := strings.ToLower(strings.TrimSuffix(r.Host, "."))
		out[h] = true
	}
	return out
}