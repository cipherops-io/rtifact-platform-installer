package testutil

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

// SandboxLayer copies infra into t.TempDir() and returns the stack layer path
// plus an absolute path to the core-network state fixture for upstream_state vars.
func SandboxLayer(t *testing.T, cloud, layer string) (layerDir, coreNetworkFixture string) {
	t.Helper()
	srcRoot := InfraRoot()
	if _, err := os.Stat(srcRoot); err != nil {
		t.Fatalf("infra root %s: %v", srcRoot, err)
	}
	tmp := t.TempDir()
	dstRoot := filepath.Join(tmp, "infra")
	if err := copyTree(srcRoot, dstRoot); err != nil {
		t.Fatalf("copy infra: %v", err)
	}
	fixtureDst := filepath.Join(tmp, "terratest", "testdata", "core-network.tfstate")
	if err := os.MkdirAll(filepath.Dir(fixtureDst), 0o755); err != nil {
		t.Fatalf("fixture dir: %v", err)
	}
	if err := copyFile(CoreNetworkStateFixture(), fixtureDst); err != nil {
		t.Fatalf("copy fixture: %v", err)
	}
	return filepath.Join(dstRoot, "stacks", "aios-infra", cloud, layer), fixtureDst
}

func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if info.IsDir() {
			if info.Name() == ".terraform" {
				return filepath.SkipDir
			}
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		return copyFile(path, filepath.Join(dst, rel))
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}