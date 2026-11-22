package deploy

import (
	"fmt"
	"regexp"
	"strings"
)

var envPlaceholderRE = regexp.MustCompile(`\$\{env\.[A-Z][A-Z0-9_]+\}`)

// MaskEnvPlaceholders redacts ${env.NAME} references in text shown to the user.
func MaskEnvPlaceholders(s string) string {
	return envPlaceholderRE.ReplaceAllString(s, "***")
}

// MaskEnvInValue returns a copy of v with string leaves masked (for explain output).
func MaskEnvInValue(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[k] = MaskEnvInValue(val)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, val := range t {
			out[i] = MaskEnvInValue(val)
		}
		return out
	case string:
		return MaskEnvPlaceholders(t)
	default:
		return v
	}
}

// ExplainSummary returns a human-readable, secret-safe summary of the deploy file.
func ExplainSummary(doc *Document) string {
	order, err := InstallOrder(doc.Releases)
	if err != nil {
		return fmt.Sprintf("(install order unavailable: %v)", err)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "kind=%s name_prefix=%s environment=%s\n", doc.Kind, doc.NamePrefix, doc.Environment)
	fmt.Fprintf(&b, "install order: %s\n", strings.Join(order, " → "))
	for _, rel := range doc.Releases {
		fmt.Fprintf(&b, "\nrelease %q (ns=%s chart=%s@%s crds=%s", rel.Name, rel.Namespace, rel.Chart, rel.Version, rel.CRDs)
		if rel.Protect {
			b.WriteString(", protect")
		}
		if len(rel.Requires) > 0 {
			fmt.Fprintf(&b, ", requires=%v", rel.Requires)
		}
		b.WriteString(")\n")
		if len(rel.Values) > 0 {
			masked := MaskEnvInValue(rel.Values)
			fmt.Fprintf(&b, "  values: %v\n", masked)
		}
	}
	return strings.TrimSpace(b.String())
}