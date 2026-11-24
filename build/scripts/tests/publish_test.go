package tests

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

// TestPublishZot runs publish-zot.sh against a throwaway Zot. It needs RTIFACT_TEST_ZOT_BIN (a Zot registry
// binary from github.com/project-zot/zot releases) plus oras and cosign on PATH, and skips otherwise.
func TestPublishZot(t *testing.T) {
	zot := os.Getenv("RTIFACT_TEST_ZOT_BIN")
	if zot == "" || testing.Short() {
		t.Skip("set RTIFACT_TEST_ZOT_BIN to a Zot registry binary to run")
	}
	for _, tool := range []string{"oras", "cosign", "jq"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not on PATH", tool)
		}
	}
	tmp := t.TempDir()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	cfg := fmt.Sprintf(`{"distSpecVersion":"1.1.1","storage":{"rootDirectory":%q,"dedupe":false},"http":{"address":"127.0.0.1","port":"%d"},"log":{"level":"error"}}`,
		filepath.Join(tmp, "data"), port)
	writeFile(t, filepath.Join(tmp, "zot.json"), cfg, 0o644)
	srv := exec.Command(zot, "serve", filepath.Join(tmp, "zot.json"))
	if err := srv.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = srv.Process.Kill(); _, _ = srv.Process.Wait() })
	registry := fmt.Sprintf("127.0.0.1:%d", port)
	for i := 0; ; i++ {
		if resp, err := http.Get("http://" + registry + "/v2/"); err == nil {
			resp.Body.Close()
			break
		}
		if i > 50 {
			t.Fatal("zot did not start")
		}
		time.Sleep(100 * time.Millisecond)
	}

	keyDir := filepath.Join(tmp, "key")
	_ = os.MkdirAll(keyDir, 0o755)
	gen := exec.Command("cosign", "generate-key-pair")
	gen.Dir = keyDir
	gen.Env = append(os.Environ(), "COSIGN_PASSWORD=test")
	if out, err := gen.CombinedOutput(); err != nil {
		t.Fatalf("cosign generate-key-pair: %v\n%s", err, out)
	}
	key, _ := os.ReadFile(filepath.Join(keyDir, "cosign.key"))

	dist := filepath.Join(tmp, "dist")
	clouds := []string{"aws", "gcp"}
	platforms := []string{"linux/amd64", "darwin/arm64"}
	build := func(gen string) {
		for _, c := range clouds {
			for _, p := range platforms {
				writeFile(t, filepath.Join(dist, "rtifact-"+c+"-"+strings.ReplaceAll(p, "/", "-")), c+" "+p+" "+gen, 0o755)
			}
		}
	}
	publish := func(extra ...string) (string, error) {
		env := []string{
			"DIST_DIR=" + dist, "REGISTRY=" + registry, "PLAIN_HTTP=true", "REPOSITORY=rtifact-installer",
			"VERSION=v1.2.3", "CLOUDS=" + strings.Join(clouds, ","), "PLATFORMS=" + strings.Join(platforms, ","),
			"COSIGN_PASSWORD=test", "COSIGN_PRIVATE_KEY=" + string(key),
			"COSIGN_PUBLIC_KEY_FILE=" + filepath.Join(keyDir, "cosign.pub"),
			"SUMMARY_FILE=" + filepath.Join(tmp, "summary.md"),
		}
		return run(t, append(env, extra...), "publish-zot.sh")
	}
	pull := func(ref, platform string) string {
		out := filepath.Join(tmp, "pull-"+strings.NewReplacer("/", "-", ":", "-").Replace(ref+platform))
		args := []string{"pull", "--plain-http", "-o", out, registry + "/rtifact-installer:" + ref}
		if platform != "" {
			args = append(args, "--platform", platform)
		}
		if b, err := exec.Command("oras", args...).CombinedOutput(); err != nil {
			t.Fatalf("oras %v: %v\n%s", args, err, b)
		}
		b, _ := os.ReadFile(filepath.Join(out, "rtifact"))
		return string(b)
	}

	build("one")
	if out, err := publish(); err != nil {
		t.Fatalf("first publish: %v\n%s", err, out)
	} else if strings.Count(out, "pushed:") != 7 || strings.Count(out, "signed:") != 7 {
		t.Fatalf("want 4 artifacts, 2 indexes, 1 checksums pushed and signed:\n%s", out)
	}
	if got := pull("v1.2.3-gcp", "darwin/arm64"); got != "gcp darwin/arm64 one" {
		t.Errorf("index pull for darwin/arm64 = %q", got)
	}
	if got := pull("latest-aws", "linux/amd64"); got != "aws linux/amd64 one" {
		t.Errorf("latest-aws pull = %q", got)
	}

	if out, err := publish(); err != nil {
		t.Fatalf("second publish: %v\n%s", err, out)
	} else if strings.Contains(out, "pushed:") || strings.Contains("\n"+out, "\nsigned:") || strings.Count(out, "already present:") != 7 {
		t.Fatalf("second publish changed something:\n%s", out)
	}

	build("two")
	if out, err := publish(); err == nil || !strings.Contains(out, "refusing to publish") {
		t.Fatalf("want refusal for changed content, got err=%v\n%s", err, out)
	}
	if got := pull("v1.2.3-aws-linux-amd64", ""); got != "aws linux/amd64 one" {
		t.Errorf("refused publish still changed a tag: %q", got)
	}

	if out, err := publish("REPLACE_EXISTING=true"); err != nil || strings.Count(out, "replaced:") != 7 {
		t.Fatalf("replace: err=%v\n%s", err, out)
	}
	if got := pull("latest-aws", "linux/amd64"); got != "aws linux/amd64 two" {
		t.Errorf("latest-aws after replace = %q", got)
	}
	verify := exec.Command("cosign", "verify", "--allow-http-registry", "--allow-insecure-registry", "--insecure-ignore-tlog=true",
		"--key", filepath.Join(keyDir, "cosign.pub"), registry+"/rtifact-installer:v1.2.3-aws")
	if b, err := verify.CombinedOutput(); err != nil {
		t.Errorf("cosign verify index: %v\n%s", err, b)
	}

	host := runtime.GOOS + "/" + runtime.GOARCH
	if !slices.Contains(platforms, host) {
		t.Logf("install.sh checks skipped: no fake artifact for %s", host)
		return
	}
	installer, err := filepath.Abs("../../../scripts/install.sh")
	if err != nil {
		t.Fatal(err)
	}
	install := func(path string, args ...string) (string, string, error) {
		dir := filepath.Join(tmp, fmt.Sprintf("install-%d", len(args))+strings.Join(args, "-"))
		cmd := exec.Command("/bin/sh", append([]string{installer, "--dir", dir}, args...)...)
		cmd.Env = append(os.Environ(), "RTIFACT_REGISTRY="+registry, "RTIFACT_REGISTRY_PLAIN_HTTP=true")
		if path != "" {
			cmd.Env = append(cmd.Env, "PATH="+path)
		}
		out, err := cmd.CombinedOutput()
		b, _ := os.ReadFile(filepath.Join(dir, "rtifact"))
		return string(out), string(b), err
	}
	want := "aws " + host + " two"
	if out, got, err := install("", "--cloud", "aws", "--key", filepath.Join(keyDir, "cosign.pub")); err != nil || got != want ||
		!strings.Contains(out, "Signature verified") {
		t.Errorf("install.sh with oras, latest: err=%v got=%q\n%s", err, got, out)
	}
	// /usr/bin:/bin has curl but not oras, so these take the registry API path.
	if out, got, err := install("/usr/bin:/bin", "--cloud", "aws"); err != nil || got != want || !strings.Contains(out, "with curl") {
		t.Errorf("install.sh with curl, latest: err=%v got=%q\n%s", err, got, out)
	}
	if out, got, err := install("/usr/bin:/bin", "--cloud", "gcp", "--version", "v1.2.3"); err != nil || got != "gcp "+host+" two" {
		t.Errorf("install.sh with curl, version: err=%v got=%q\n%s", err, got, out)
	}
	if out, _, err := install("/usr/bin:/bin", "--cloud", "aws", "--version", "v9.9.9"); err == nil || !strings.Contains(out, "no artifact") {
		t.Errorf("install.sh missing version: want failure, err=%v\n%s", err, out)
	}
	if out, _, err := install("", "--cloud", "oci"); err == nil || !strings.Contains(out, "unknown cloud") {
		t.Errorf("install.sh unknown cloud: want failure, err=%v\n%s", err, out)
	}
}
