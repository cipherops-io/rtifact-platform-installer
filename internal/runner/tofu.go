package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cipherops-io/rtifact-platform-installer/internal/version"
)

func resolveTofu(contentRoot string) (tofuPath, mirrorDir string, cleanup func(), err error) {
	embedTofu := filepath.Join(contentRoot, "build", "embed", "tofu", runtime.GOOS+"_"+runtime.GOARCH, "tofu")
	if runtime.GOOS == "windows" {
		embedTofu += ".exe"
	}
	mirrorDir = providerMirrorDir(contentRoot)
	if st, err := os.Stat(embedTofu); err == nil && !st.IsDir() {
		if version.IsRelease() && mirrorDir == "" {
			return "", "", nil, fmt.Errorf("release build is missing provider mirror under build/embed/mirror; run make embed before make release")
		}
		return embedTofu, mirrorDir, nil, nil
	}
	if version.IsRelease() {
		return "", "", nil, fmt.Errorf("release build is missing embedded OpenTofu under build/embed/tofu; rebuild with make release after make embed")
	}

	if p, err := exec.LookPath("tofu"); err == nil {
		if base := filepath.Base(p); strings.HasPrefix(base, "terraform") {
			return "", "", nil, fmt.Errorf("refusing to run %q: install OpenTofu (tofu), not HashiCorp Terraform", p)
		}
		return p, mirrorDir, nil, nil
	}

	return "", "", nil, fmt.Errorf("OpenTofu (tofu) not found on PATH; install https://opentofu.org/docs/intro/install/ or run make fetch-tofu")
}

// providerMirrorDir returns the filesystem provider mirror path when populated (release/offline).
func providerMirrorDir(contentRoot string) string {
	p := filepath.Join(contentRoot, "build", "embed", "mirror")
	st, err := os.Stat(p)
	if err != nil || !st.IsDir() {
		return ""
	}
	entries, err := os.ReadDir(p)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		switch e.Name() {
		case "placeholder.txt", ".keep", "README.txt":
			continue
		default:
			return p
		}
	}
	return ""
}

func writeProviderRC(path, mirror string) error {
	body := fmt.Sprintf(`provider_installation {
  filesystem_mirror {
    path = %q
  }
  direct {
    exclude = ["*/*/*"]
  }
}
`, mirror)
	return os.WriteFile(path, []byte(body), 0o644)
}

// pluginCacheDir is reused across stack runs so providers are not re-downloaded for every temp work dir.
func pluginCacheDir(contentRoot string) (string, error) {
	pc, err := filepath.Abs(filepath.Join(contentRoot, "build", ".plugin-cache"))
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(pc, 0o755); err != nil {
		return "", err
	}
	return pc, nil
}

const tofuInitAttempts = 3

func runTofuInit(tofuPath, stackWork string, env []string, initArgs []string) error {
	var lastErr error
	backoff := 2 * time.Second
	for attempt := 1; attempt <= tofuInitAttempts; attempt++ {
		if attempt > 1 {
			time.Sleep(backoff)
			backoff *= 2
		}
		var stderr bytes.Buffer
		c := exec.Command(tofuPath, initArgs...)
		c.Dir = stackWork
		c.Env = env
		c.Stdout = os.Stdout
		c.Stderr = io.MultiWriter(os.Stderr, &stderr)
		err := c.Run()
		if err == nil {
			return nil
		}
		lastErr = err
		msg := stderr.String()
		if attempt == tofuInitAttempts || !retryableTofuInit(stderr.String()) {
			return err
		}
		fmt.Fprintf(os.Stderr, "OpenTofu init failed (attempt %d/%d); retrying after %s…\n", attempt, tofuInitAttempts, backoff)
		if msg != "" && strings.Contains(msg, "connection reset") {
			fmt.Fprintf(os.Stderr, "  (provider registry connection was reset)\n")
		}
	}
	return lastErr
}

func retryableTofuInit(stderr string) bool {
	retryHints := []string{
		"Failed to install provider",
		"connection reset",
		"connection refused",
		"i/o timeout",
		"temporary failure",
		"TLS handshake timeout",
		"unexpected EOF",
		"Error while installing",
	}
	lower := strings.ToLower(stderr)
	for _, h := range retryHints {
		if strings.Contains(lower, strings.ToLower(h)) {
			return true
		}
	}
	return false
}
