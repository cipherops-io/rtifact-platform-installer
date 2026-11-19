package runner

import (
	"os"
	"path/filepath"
	"strings"
)

// chmodProviderBinaries ensures provider plugin files are executable (macOS/Linux).
func chmodProviderBinaries(stackWork, pluginCache string) error {
	roots := []string{filepath.Join(stackWork, ".terraform", "providers")}
	if pluginCache != "" {
		roots = append(roots, pluginCache)
	}
	for _, root := range roots {
		if _, err := os.Stat(root); err != nil {
			continue
		}
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			name := d.Name()
			if strings.HasPrefix(name, "terraform-provider-") {
				return os.Chmod(path, 0o755)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
