package deploy_test

import (
	"path/filepath"
	"runtime"
	"testing"
)

// repoRoot is the rtifact-platform-installer checkout (tests/deploy → ../..).
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}