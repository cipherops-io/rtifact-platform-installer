package runner

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

func injectDNSDelegation(ctx context.Context, layer stacks.Layer, customer *config.CustomerConfig, merged map[string]any) {
	if layer != stacks.LayerNetworkEdge || customer.Cloud != "aws" {
		return
	}
	if dnsDelegationVerified(ctx, customer.NamePrefix(), merged) {
		merged["dns_delegation_verified"] = true
	}
}

func dnsDelegationVerified(ctx context.Context, prefix string, merged map[string]any) bool {
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
	expected := readSavedNameServers(prefix)
	live, err := net.DefaultResolver.LookupNS(ctx, zone)
	if err != nil || len(live) == 0 {
		return false
	}
	liveSet := nsHostSet(live)
	if len(expected) > 0 {
		for _, e := range expected {
			if !liveSet[strings.ToLower(strings.TrimSuffix(e, "."))] {
				return false
			}
		}
		return true
	}
	return len(liveSet) >= 2
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