package deploy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
)

func TestDiscover_cwdKubeconfig(t *testing.T) {
	dir := t.TempDir()
	prefix := "rtfc-sbx2"
	kc := filepath.Join(dir, prefix+"-kubeconfig.yaml")
	if err := os.WriteFile(kc, []byte("apiVersion: v1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	doc := &Document{NamePrefix: prefix}
	oldwd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldwd) }()

	d, err := Discover(DiscoverOpts{Doc: doc, FilePath: filepath.Join(dir, "releases.yaml")})
	if err != nil {
		t.Fatal(err)
	}
	same := d.KubeconfigPath == kc
	if !same {
		a, err1 := filepath.EvalSymlinks(d.KubeconfigPath)
		b, err2 := filepath.EvalSymlinks(kc)
		same = err1 == nil && err2 == nil && a == b
	}
	if !same {
		t.Fatalf("got %q want %q", d.KubeconfigPath, kc)
	}
}

func TestDiscover_kubeconfigFlagRelativeCWD(t *testing.T) {
	dir := t.TempDir()
	kc := filepath.Join(dir, "foo-kubeconfig.yaml")
	if err := os.WriteFile(kc, []byte("apiVersion: v1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(dir, "cfg")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	doc := &Document{NamePrefix: "pfx"}
	oldwd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldwd) }()

	d, err := Discover(DiscoverOpts{
		Doc:           doc,
		FilePath:      filepath.Join(sub, "releases.yaml"),
		KubeconfigCLI: "foo-kubeconfig.yaml",
		WorkDir:       dir,
	})
	if err != nil {
		t.Fatal(err)
	}
	same := d.KubeconfigPath == kc
	if !same {
		a, e1 := filepath.EvalSymlinks(d.KubeconfigPath)
		b, e2 := filepath.EvalSymlinks(kc)
		same = e1 == nil && e2 == nil && a == b
	}
	if !same {
		t.Fatalf("flag path relative to cwd: got %q want %q", d.KubeconfigPath, kc)
	}
}

func TestDiscover_missing(t *testing.T) {
	doc := &Document{NamePrefix: "missing-prefix"}
	_, err := Discover(DiscoverOpts{Doc: doc, FilePath: "releases.yaml", WorkDir: t.TempDir()})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestIdentityGuardRefusal_message(t *testing.T) {
	if err := IdentityGuardRefusal("https://a", "https://b"); err == nil {
		t.Fatal("expected error")
	}
	_ = config.ArtifactsDir("x")
}