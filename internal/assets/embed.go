package assets

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/packaged"
)

// ContentRootOptional is ContentRoot without treating missing layout as fatal for optional callers.
func ContentRootOptional() (string, bool, error) {
	return ContentRoot()
}

// ContentRoot returns the directory containing profiles/ and infra/.
func ContentRoot() (string, bool, error) {
	if root := os.Getenv("INSTALLER_CONTENT_ROOT"); root != "" {
		if ok := hasContentLayout(root); ok {
			return root, false, nil
		}
		return "", false, fs.ErrNotExist
	}

	if cwd, err := os.Getwd(); err == nil {
		if root, ok := findContentRoot(cwd); ok {
			return root, false, nil
		}
	}

	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		if root, ok := findContentRoot(dir); ok {
			return root, false, nil
		}
	}

	source := packaged.FS()
	if entries, err := fs.ReadDir(source, "profiles"); err == nil && len(entries) > 0 {
		tmp, err := extractPackagedToTemp(source)
		if err != nil {
			return "", false, err
		}
		if hasContentLayout(tmp) {
			return tmp, true, nil
		}
		_ = os.RemoveAll(tmp)
	}

	return "", false, fs.ErrNotExist
}

func hasContentLayout(root string) bool {
	_, err1 := os.Stat(filepath.Join(root, "profiles"))
	_, err2 := os.Stat(filepath.Join(root, "infra"))
	return err1 == nil && err2 == nil
}

func findContentRoot(start string) (string, bool) {
	dir := start
	for i := 0; i < 8; i++ {
		if hasContentLayout(dir) {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", false
}

func extractPackagedToTemp(source fs.FS) (string, error) {
	tmp, err := os.MkdirTemp("", "installer-content-")
	if err != nil {
		return "", err
	}
	err = fs.WalkDir(source, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		dst := filepath.Join(tmp, path)
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		data, err := fs.ReadFile(source, path)
		if err != nil {
			return err
		}
		perm := os.FileMode(0o644)
		if embeddedExecutable(path) {
			perm = 0o755
		}
		return os.WriteFile(dst, data, perm)
	})
	if err != nil {
		os.RemoveAll(tmp)
		return "", err
	}
	return tmp, nil
}

func embeddedExecutable(embedPath string) bool {
	base := filepath.Base(embedPath)
	if base != "tofu" && base != "tofu.exe" {
		return false
	}
	return strings.Contains(embedPath, "build/embed/tofu/")
}

// ReadSchemaFile reads customer-config.schema.json from content root or embedded schemas.
func ReadSchemaFile(name string) ([]byte, error) {
	root, _, err := ContentRoot()
	if err == nil {
		p := filepath.Join(root, "schemas", name)
		if b, err := os.ReadFile(p); err == nil {
			return b, nil
		}
	}
	return fs.ReadFile(packaged.FS(), filepath.Join("schemas", name))
}