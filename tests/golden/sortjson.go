package golden

import (
	"encoding/json"
	"sort"
)

func sortValue(v any) any {
	switch t := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		m := make(map[string]any, len(t))
		for _, k := range keys {
			m[k] = sortValue(t[k])
		}
		return m
	case []any:
		out := make([]any, len(t))
		for i, val := range t {
			out[i] = sortValue(val)
		}
		return out
	default:
		return v
	}
}

func marshalSortedJSON(v any) ([]byte, error) {
	b, err := json.MarshalIndent(sortValue(v), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}