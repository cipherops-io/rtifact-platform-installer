package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
)

// WriteInitConfig copies the packaged minimal deploy example to out.
func WriteInitConfig(out string) error {
	root, _, err := assets.ContentRoot()
	if err != nil {
		return err
	}
	src := filepath.Join(root, "examples", "deploy", "minimal.yaml")
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("examples/deploy/minimal.yaml: %w", err)
	}
	dir := filepath.Dir(out)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(out, data, 0o644)
}