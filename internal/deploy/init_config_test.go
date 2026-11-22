package deploy

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWriteInitConfig_force(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	t.Setenv("INSTALLER_CONTENT_ROOT", root)
	dir := t.TempDir()
	out := filepath.Join(dir, "releases.yaml")
	if err := WriteInitConfig(out, false); err != nil {
		t.Fatal(err)
	}
	if err := WriteInitConfig(out, false); err == nil {
		t.Fatal("expected exists error")
	}
	if err := WriteInitConfig(out, true); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(out)
	if err != nil || st.Size() == 0 {
		t.Fatalf("stat: %v", err)
	}
}