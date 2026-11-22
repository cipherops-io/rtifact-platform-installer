package deploy_test

import (
	"path/filepath"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy"
)

func TestDeployValidateExample(t *testing.T) {
	root := repoRoot(t)
	t.Setenv("INSTALLER_CONTENT_ROOT", root)
	path := filepath.Join(root, "examples", "deploy", "minimal.yaml")
	if err := deploy.ValidateFile(path); err != nil {
		t.Fatal(err)
	}
}