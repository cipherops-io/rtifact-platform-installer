package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func WriteMergedConfigArtifact(prefix, layer string, merged map[string]any) error {
	dir := filepath.Join(os.Getenv("HOME"), ".rtifact-installer", prefix)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return err
	}
	name := fmt.Sprintf("merged-config-%s.json", layer)
	return os.WriteFile(filepath.Join(dir, name), b, 0o644)
}