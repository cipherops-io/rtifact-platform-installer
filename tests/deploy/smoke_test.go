package deploy_test

import (
	"path/filepath"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy"
)

func TestDeployValidateExample(t *testing.T) {
	root := repoRoot(t)
	t.Setenv("INSTALLER_CONTENT_ROOT", root)
	for _, rel := range []string{
		"examples/deploy/minimal.yaml",
		"examples/deploy/demo/demo.yaml",
	} {
		path := filepath.Join(root, rel)
		if err := deploy.ValidateFile(path); err != nil {
			t.Fatalf("%s: %v", rel, err)
		}
	}
}