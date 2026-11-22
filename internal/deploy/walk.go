package deploy

// walkYAML applies fn to every string leaf in a YAML-decoded tree.
func walkYAML(v any, fn func(string) error) error {
	switch t := v.(type) {
	case map[string]any:
		for _, val := range t {
			if err := walkYAML(val, fn); err != nil {
				return err
			}
		}
	case []any:
		for _, item := range t {
			if err := walkYAML(item, fn); err != nil {
				return err
			}
		}
	case string:
		return fn(t)
	}
	return nil
}

// transformYAMLStrings returns a copy of v with string leaves rewritten by fn.
func transformYAMLStrings(v any, fn func(string) string) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[k] = transformYAMLStrings(val, fn)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, item := range t {
			out[i] = transformYAMLStrings(item, fn)
		}
		return out
	case string:
		return fn(t)
	default:
		return v
	}
}