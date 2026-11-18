package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
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

func writeBackend(stackDir string, b *config.BackendBlock, az config.AzureBlock) error {
	if b.Type != "azurerm" {
		return fmt.Errorf("unsupported backend type %q", b.Type)
	}
	tenant := az.TenantID
	authLine := ""
	if tenant != "" {
		authLine = fmt.Sprintf("    tenant_id           = %q\n    use_azuread_auth    = true\n", tenant)
	}
	content := fmt.Sprintf(`terraform {
  backend "azurerm" {
    resource_group_name  = %q
    storage_account_name = %q
    container_name       = %q
    key                  = %q
    subscription_id      = %q
%s  }
}
`, b.ResourceGroupName, b.StorageAccountName, b.ContainerName, b.Key, az.SubscriptionID, authLine)
	return os.WriteFile(filepath.Join(stackDir, "backend.tofu"), []byte(content), 0o644)
}