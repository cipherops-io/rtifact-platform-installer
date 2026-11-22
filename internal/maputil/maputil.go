package maputil

import "fmt"

// DeepCopyMap returns a shallow tree copy of nested maps.
func DeepCopyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		switch t := v.(type) {
		case map[string]any:
			out[k] = DeepCopyMap(t)
		case map[interface{}]interface{}:
			out[k] = YAMLMapToAnyMap(t)
		default:
			out[k] = v
		}
	}
	return out
}

// YAMLMapToAnyMap converts yaml.v3 map keys to string keys.
func YAMLMapToAnyMap(in map[interface{}]interface{}) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		key, _ := k.(string)
		switch t := v.(type) {
		case map[interface{}]interface{}:
			out[key] = YAMLMapToAnyMap(t)
		default:
			out[key] = v
		}
	}
	return out
}

// Truthy is true only for boolean true.
func Truthy(v any) bool {
	b, ok := v.(bool)
	return ok && b
}

// StringValue extracts a string from YAML-decoded values.
func StringValue(v any) string {
	s, _ := v.(string)
	return s
}

// StringSlice converts a YAML list to []string.
func StringSlice(v any) ([]string, error) {
	arr, ok := v.([]any)
	if !ok {
		return nil, nil
	}
	out := make([]string, 0, len(arr))
	for _, x := range arr {
		s, ok := x.(string)
		if !ok {
			return nil, nil
		}
		out = append(out, s)
	}
	return out, nil
}

// FlattenStringSlice extracts non-empty strings from a []any slice.
func FlattenStringSlice(v any) []string {
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

// SliceOfMaps converts a YAML/JSON list of mappings to []map[string]any.
func SliceOfMaps(v any) ([]map[string]any, error) {
	switch arr := v.(type) {
	case []any:
		return sliceOfMapsFromSlice(arr)
	case []map[string]any:
		return arr, nil
	default:
		return nil, fmt.Errorf("expected a list")
	}
}

func sliceOfMapsFromSlice(arr []any) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(arr))
	for i, x := range arr {
		m, ok := x.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("entry %d is not a mapping", i)
		}
		out = append(out, m)
	}
	return out, nil
}

// FirstString returns the first non-empty string value for keys in m.
func FirstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if s := StringValue(m[k]); s != "" {
			return s
		}
	}
	return ""
}