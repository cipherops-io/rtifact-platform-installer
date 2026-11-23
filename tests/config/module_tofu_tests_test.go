package config_test

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestModuleTofuTests runs `tofu test` (mocked providers) for every module with a tests/ directory.
// Each module is copied to a temp dir so init never writes lock files or .terraform into the repo.
func TestModuleTofuTests(t *testing.T) {
	tofu, err := exec.LookPath("tofu")
	if err != nil {
		t.Skip("tofu not on PATH")
	}
	if testing.Short() {
		t.Skip("short mode")
	}
	modules, err := filepath.Glob(filepath.Join("..", "..", "infra", "modules", "*", "*", "tests"))
	if err != nil {
		t.Fatal(err)
	}
	if len(modules) == 0 {
		t.Fatal("no module tests found")
	}
	cache := t.TempDir()
	for _, testsDir := range modules {
		src := filepath.Dir(testsDir)
		name := filepath.Base(filepath.Dir(src)) + "/" + filepath.Base(src)
		t.Run(name, func(t *testing.T) {
			dst := t.TempDir()
			if err := copyModule(src, dst); err != nil {
				t.Fatal(err)
			}
			env := append(os.Environ(), "TF_PLUGIN_CACHE_DIR="+cache, "TF_IN_AUTOMATION=1")
			for _, args := range [][]string{{"init", "-backend=false", "-input=false", "-no-color"}, {"test", "-no-color"}} {
				cmd := exec.Command(tofu, args...)
				cmd.Dir = dst
				cmd.Env = env
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("tofu %s: %v\n%s", strings.Join(args, " "), err, out)
				}
			}
		})
	}
}

func copyModule(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".terraform" {
				return filepath.SkipDir
			}
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dst, rel), b, 0o644)
	})
}
