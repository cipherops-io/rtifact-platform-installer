package deploy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/paths"
)

func TestExpandHomePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	got, err := expandHomePath("~/foo/kube.yaml")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, "foo", "kube.yaml")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if _, err := expandHomePath("/abs/path"); err != nil {
		t.Fatal(err)
	}
}

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

func TestDiscover_flagWinsOverFile(t *testing.T) {
	dir := t.TempDir()
	flagKC := filepath.Join(dir, "from-flag.yaml")
	fileKC := filepath.Join(dir, "from-file.yaml")
	if err := os.WriteFile(flagKC, []byte("apiVersion: v1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fileKC, []byte("apiVersion: v1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	doc := &Document{
		NamePrefix: "pfx",
		Cluster:    Cluster{Kubeconfig: "from-file.yaml"},
	}
	d, err := Discover(DiscoverOpts{
		Doc:           doc,
		FilePath:      filepath.Join(dir, "releases.yaml"),
		KubeconfigCLI: flagKC,
		WorkDir:       dir,
	})
	if err != nil {
		t.Fatal(err)
	}
	same := d.KubeconfigPath == flagKC
	if !same {
		a, e1 := filepath.EvalSymlinks(d.KubeconfigPath)
		b, e2 := filepath.EvalSymlinks(flagKC)
		same = e1 == nil && e2 == nil && a == b
	}
	if !same {
		t.Fatalf("flag should win: got %q", d.KubeconfigPath)
	}
}

func TestDiscover_artifactKubeconfig(t *testing.T) {
	dir := t.TempDir()
	prefix := "art-pfx"
	artifact := paths.ArtifactsDir(prefix)
	if err := os.MkdirAll(artifact, 0o700); err != nil {
		t.Fatal(err)
	}
	kc := filepath.Join(artifact, prefix+"-kubeconfig.yaml")
	if err := os.WriteFile(kc, []byte("apiVersion: v1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	doc := &Document{NamePrefix: prefix}
	d, err := Discover(DiscoverOpts{
		Doc:      doc,
		FilePath: filepath.Join(dir, "releases.yaml"),
		WorkDir:  dir,
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
		t.Fatalf("artifact kubeconfig: got %q want %q", d.KubeconfigPath, kc)
	}
	if d.KubeconfigFrom != "discovered in artifact directory" {
		t.Fatalf("from: %q", d.KubeconfigFrom)
	}
}

func TestDiscover_infraOutputs(t *testing.T) {
	dir := t.TempDir()
	prefix := "pfx"
	if err := os.WriteFile(filepath.Join(dir, prefix+"-kubeconfig.yaml"), []byte("apiVersion: v1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	outs := filepath.Join(dir, "outputs-dir")
	if err := os.MkdirAll(outs, 0o755); err != nil {
		t.Fatal(err)
	}
	doc := &Document{NamePrefix: prefix, InfraOutputs: "outputs-dir"}
	d, err := Discover(DiscoverOpts{
		Doc:      doc,
		FilePath: filepath.Join(dir, "releases.yaml"),
		WorkDir:  dir,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "outputs-dir")
	same := d.OutputsPath == want
	if !same {
		a, e1 := filepath.EvalSymlinks(d.OutputsPath)
		b, e2 := filepath.EvalSymlinks(want)
		same = e1 == nil && e2 == nil && a == b
	}
	if !same {
		t.Fatalf("infra_outputs file field: got %q", d.OutputsPath)
	}
	if d.OutputsFrom != "infra_outputs in file" {
		t.Fatalf("from: %q", d.OutputsFrom)
	}

	prefix2 := "disc-pfx"
	if err := os.WriteFile(filepath.Join(dir, prefix2+"-kubeconfig.yaml"), []byte("apiVersion: v1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	artifact := paths.ArtifactsDir(prefix2)
	if err := os.MkdirAll(artifact, 0o700); err != nil {
		t.Fatal(err)
	}
	doc2 := &Document{NamePrefix: prefix2}
	d2, err := Discover(DiscoverOpts{
		Doc:      doc2,
		FilePath: filepath.Join(dir, "rel.yaml"),
		WorkDir:  dir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if d2.OutputsPath != artifact {
		t.Fatalf("discovered outputs: got %q want %q", d2.OutputsPath, artifact)
	}
}

func TestIdentityGuardRefusal_message(t *testing.T) {
	if err := IdentityGuardRefusal("https://a", "https://b"); err == nil {
		t.Fatal("expected error")
	}
	_ = paths.ArtifactsDir("x")
}