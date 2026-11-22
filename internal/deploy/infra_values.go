package deploy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func resolveInfraKey(outputsDir, key string) (string, error) {
	dir := outputsDir
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read infra outputs %s: %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() || !jsonExt(e.Name()) {
			continue
		}
		path := filepath.Join(dir, e.Name())
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var flat map[string]any
		if json.Unmarshal(b, &flat) != nil {
			continue
		}
		if v, ok := flat[key]; ok {
			if s, ok := v.(string); ok {
				return s, nil
			}
			return fmt.Sprint(v), nil
		}
	}
	return "", fmt.Errorf("${infra.%s} not found under %s", key, dir)
}

func jsonExt(name string) bool {
	return len(name) > 6 && name[len(name)-5:] == ".json"
}