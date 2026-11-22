package maputil

import "fmt"

// NormalizeYAML converts yaml.v3 decode trees into JSON-schema-friendly values.
func NormalizeYAML(v any) any {
	switch t := v.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]any, len(t))
		for k, val := range t {
			m[fmt.Sprint(k)] = NormalizeYAML(val)
		}
		return m
	case map[string]any:
		m := make(map[string]any, len(t))
		for k, val := range t {
			m[k] = NormalizeYAML(val)
		}
		return m
	case []interface{}:
		s := make([]any, len(t))
		for i, val := range t {
			s[i] = NormalizeYAML(val)
		}
		return s
	default:
		return v
	}
}