package deploy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LookupInfraScalar resolves ${infra.key} from *-outputs.json files under outputsDir.
// Keys may be layer-qualified (layer.output_key) or a bare output_key unique across layers.
func LookupInfraScalar(outputsDir, key string) (string, error) {
	if outputsDir == "" {
		return "", fmt.Errorf("${infra.%s} requires infra_outputs or discovered outputs", key)
	}
	if dot := strings.Index(key, "."); dot > 0 {
		layer := key[:dot]
		outKey := key[dot+1:]
		if outKey == "" {
			return "", fmt.Errorf("${infra.%s}: invalid qualified key", key)
		}
		return lookupInLayerFile(outputsDir, layer, outKey)
	}
	return lookupBareKey(outputsDir, key)
}

func lookupInLayerFile(dir, layer, key string) (string, error) {
	path := filepath.Join(dir, layer+"-outputs.json")
	flat, err := readOutputsFlat(path)
	if err != nil {
		return "", fmt.Errorf("${infra.%s.%s}: %w", layer, key, err)
	}
	return scalarFromFlat(flat, key, fmt.Sprintf("${infra.%s.%s}", layer, key))
}

func lookupBareKey(dir, key string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read infra outputs %s: %w", dir, err)
	}
	var (
		found    string
		foundIn  string
		ambiguous []string
	)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), "-outputs.json") {
			continue
		}
		layer := strings.TrimSuffix(e.Name(), "-outputs.json")
		path := filepath.Join(dir, e.Name())
		flat, err := readOutputsFlat(path)
		if err != nil {
			continue
		}
		if _, ok := flat[key]; !ok {
			continue
		}
		if foundIn != "" {
			ambiguous = append(ambiguous, layer)
			continue
		}
		val, err := scalarFromFlat(flat, key, fmt.Sprintf("${infra.%s}", key))
		if err != nil {
			return "", err
		}
		found = val
		foundIn = layer
	}
	if len(ambiguous) > 0 {
		layers := append([]string{foundIn}, ambiguous...)
		return "", fmt.Errorf("${infra.%s}: ambiguous (found in %s)", key, strings.Join(layers, ", "))
	}
	if foundIn == "" {
		return "", fmt.Errorf("${infra.%s} not found under %s", key, dir)
	}
	return found, nil
}

func readOutputsFlat(path string) (map[string]any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var flat map[string]any
	if err := json.Unmarshal(b, &flat); err != nil {
		return nil, fmt.Errorf("parse %s: %w", filepath.Base(path), err)
	}
	return flat, nil
}

func scalarFromFlat(flat map[string]any, key, ref string) (string, error) {
	v, ok := flat[key]
	if !ok {
		return "", fmt.Errorf("%s not found", ref)
	}
	switch t := v.(type) {
	case string:
		return t, nil
	case bool, float64, int, int64:
		return fmt.Sprint(t), nil
	case json.Number:
		return t.String(), nil
	default:
		return "", fmt.Errorf("%s: value must be a scalar", ref)
	}
}