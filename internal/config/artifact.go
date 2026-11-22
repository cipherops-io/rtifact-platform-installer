package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/paths"
)

func WriteMergedConfigArtifact(prefix, layer string, merged map[string]any) error {
	dir := paths.ArtifactsDir(prefix)
	b, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return err
	}
	name := fmt.Sprintf("merged-config-%s.json", layer)
	return os.WriteFile(filepath.Join(dir, name), b, 0o644)
}
