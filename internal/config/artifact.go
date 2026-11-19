package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func WriteMergedConfigArtifact(prefix, layer string, merged map[string]any) error {
	dir := ArtifactsDir(prefix)
	b, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return err
	}
	name := fmt.Sprintf("merged-config-%s.json", layer)
	return os.WriteFile(filepath.Join(dir, name), b, 0o644)
}