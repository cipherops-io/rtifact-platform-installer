package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/redact"
	"gopkg.in/yaml.v3"
)

// MergeReleaseValues reads values_files then overlays inline values for one release.
func MergeReleaseValues(deployPath string, rel Release) (map[string]any, error) {
	dir := filepath.Dir(deployPath)
	out := map[string]any{}
	for _, vf := range rel.ValuesFiles {
		p := vf
		if !filepath.IsAbs(p) {
			p = filepath.Join(dir, vf)
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		var chunk map[string]any
		if err := yaml.Unmarshal(raw, &chunk); err != nil {
			return nil, fmt.Errorf("values file %s: %w", vf, err)
		}
		out = mergeValMaps(out, chunk)
	}
	if len(rel.Values) > 0 {
		out = mergeValMaps(out, rel.Values)
	}
	return out, nil
}

func mergeValMaps(base, overlay map[string]any) map[string]any {
	if base == nil {
		base = map[string]any{}
	}
	out := make(map[string]any, len(base)+len(overlay))
	for k, v := range base {
		out[k] = v
	}
	for k, ov := range overlay {
		if bv, ok := out[k]; ok {
			bm, ok1 := bv.(map[string]any)
			om, ok2 := ov.(map[string]any)
			if ok1 && ok2 {
				out[k] = mergeValMaps(bm, om)
				continue
			}
		}
		out[k] = ov
	}
	return out
}

// ResolveReleaseValuesFrom substitutes ${env.NAME} and ${infra.key} placeholders.
func ResolveReleaseValuesFrom(tree map[string]any, rd *redact.Redactor, outputsDir string) (map[string]any, error) {
	v, err := resolveReleaseNode(tree, rd, outputsDir)
	if err != nil {
		return nil, err
	}
	out, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("resolved values must be a mapping")
	}
	return out, nil
}

func resolveReleaseNode(v any, rd *redact.Redactor, outputsDir string) (any, error) {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			nv, err := resolveReleaseNode(val, rd, outputsDir)
			if err != nil {
				return nil, err
			}
			out[k] = nv
		}
		return out, nil
	case []any:
		out := make([]any, len(t))
		for i, val := range t {
			nv, err := resolveReleaseNode(val, rd, outputsDir)
			if err != nil {
				return nil, err
			}
			out[i] = nv
		}
		return out, nil
	case string:
		return resolveReleaseString(t, rd, outputsDir)
	default:
		return v, nil
	}
}

func resolveReleaseString(s string, rd *redact.Redactor, outputsDir string) (string, error) {
	if !strings.Contains(s, "${") {
		return s, nil
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		j := strings.Index(s[i:], "${")
		if j < 0 {
			b.WriteString(s[i:])
			break
		}
		b.WriteString(s[i : i+j])
		start := i + j
		end := strings.Index(s[start:], "}")
		if end < 0 {
			b.WriteString(s[start:])
			break
		}
		token := s[start : start+end+1]
		if EnvPlaceholderToken.MatchString(token) {
			name := strings.TrimSuffix(strings.TrimPrefix(token, "${env."), "}")
			val, ok := os.LookupEnv(name)
			if !ok || val == "" {
				return "", fmt.Errorf("environment variable %s is not set (referenced as %s)", name, token)
			}
			rd.Record(val)
			b.WriteString(val)
		} else if InfraPlaceholderToken.MatchString(token) {
			key := strings.TrimSuffix(strings.TrimPrefix(token, "${infra."), "}")
			if outputsDir == "" {
				return "", fmt.Errorf("${infra.%s} requires infra_outputs or discovered outputs", key)
			}
			val, err := LookupInfraScalar(outputsDir, key)
			if err != nil {
				return "", err
			}
			b.WriteString(val)
		} else {
			b.WriteString(token)
		}
		i = start + end + 1
	}
	return b.String(), nil
}

// ResolveAllReleases resolves placeholders for every release into one redactor scope.
func ResolveAllReleases(deployPath string, doc *Document, outputsDir string, rd *redact.Redactor) (map[string]map[string]any, error) {
	out := make(map[string]map[string]any, len(doc.Releases))
	for _, rel := range doc.Releases {
		merged, err := MergeReleaseValues(deployPath, rel)
		if err != nil {
			return nil, err
		}
		resolved, err := ResolveReleaseValuesFrom(merged, rd, outputsDir)
		if err != nil {
			return nil, err
		}
		out[rel.Name] = resolved
	}
	return out, nil
}

// ValuesDigest returns a stable hash of release values (see plan.ValuesDigest).
func ValuesDigest(v map[string]any) (string, error) {
	return plan.ValuesDigest(v)
}