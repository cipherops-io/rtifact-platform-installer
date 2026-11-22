package paths

import (
	"os"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/version"
)

// ArtifactsDir returns ~/.<ProductName>/<prefix> (created on demand).
func ArtifactsDir(prefix string) string {
	dir := filepath.Join(os.Getenv("HOME"), "."+version.ProductName, prefix)
	_ = os.MkdirAll(dir, 0o755)
	return dir
}