package features

import (
	"fmt"
	"sort"
	"strings"
)

// PasteableFeaturesYAML builds a paste-ready features block from registry stable features.
func (r *Registry) PasteableFeaturesYAML() string {
	if r == nil || len(r.Features) == 0 {
		return "features:\n  # (no features registry loaded)"
	}
	var keys []string
	for name, ent := range r.Features {
		if ent.Status == "stable" {
			keys = append(keys, name)
		}
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("features:\n")
	for _, k := range keys {
		fmt.Fprintf(&b, "  %s: true\n", k)
	}
	return strings.TrimRight(b.String(), "\n")
}