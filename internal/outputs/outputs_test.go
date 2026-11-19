package outputs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

func TestMaterializeCWD(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	flat := map[string]any{
		"kubeconfig_raw":              "apiVersion: v1\nkind: Config\n",
		"bastion_ssh_private_key_pem": "-----BEGIN RSA PRIVATE KEY-----\nMIIB\n-----END RSA PRIVATE KEY-----\n",
	}
	written, err := MaterializeCWD("cm-test", flat, MaterializeOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if len(written) != 2 {
		t.Fatalf("expected 2 files, got %v", written)
	}
	for _, name := range []string{"cm-test-kubeconfig.yaml", "cm-test-bastion.pem"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
}

func TestMaterializeFromArtifacts_missing(t *testing.T) {
	got, err := MaterializeFromArtifacts("no-such-prefix-xyz", MaterializeOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("expected nil slice, got %v", got)
	}
	_ = stacks.LayerAppPlatform
}
