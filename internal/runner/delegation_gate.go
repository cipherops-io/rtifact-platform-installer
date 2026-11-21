package runner

import (
	"fmt"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
)

func truthy(v any) bool {
	b, ok := v.(bool)
	return ok && b
}

func stringValue(v any) string {
	s, _ := v.(string)
	return s
}

// delegationPauseIfNeeded reads captured network-edge outputs and pauses when NS delegation is pending.
func delegationPauseIfNeeded(cloudName string, merged map[string]any, configPath, prefix string, flat map[string]any) error {
	if cloudName != "aws" {
		return nil
	}
	features, _ := merged["features"].(map[string]any)
	if !truthy(features["public_dns"]) || !truthy(features["certificates"]) {
		return nil
	}
	pdns, _ := merged["public_dns"].(map[string]any)
	if stringValue(pdns["parent_zone_id"]) != "" || truthy(pdns["delegated"]) {
		return nil
	}
	ns := flattenStringSlice(flat["dns_name_servers"])
	if len(ns) == 0 {
		return nil
	}
	if arn := stringValue(flat["certificate_regional_arn"]); arn != "" {
		return nil
	}
	zone := stringValue(pdns["zone_name"])
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

func flattenStringSlice(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, x := range arr {
		if s, ok := x.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}