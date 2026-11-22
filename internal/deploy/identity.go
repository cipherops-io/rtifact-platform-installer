package deploy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
)

// IdentityInfo is the cluster identity check result for plan/apply banners.
type IdentityInfo struct {
	// Server is the API server from kubeconfig (normalized for display).
	Server string
	// Expected is the reference server when verified; empty when not verified.
	Expected string
	// Verified is true when kubeconfig server matched Expected.
	Verified bool
}

// Line returns the banner Identity line (no trailing newline).
func (i IdentityInfo) Line() string {
	if i.Verified {
		return fmt.Sprintf("Identity:   verified (%s)", i.Expected)
	}
	return "Identity:   not verified"
}

// CheckIdentity compares kubeconfig server to expected_server or infra outputs.
func CheckIdentity(kcfg kube.Config, doc *Document, discovery Discovery) (IdentityInfo, error) {
	info := IdentityInfo{}
	got, err := kube.ServerURL(kcfg)
	if err != nil {
		return info, err
	}
	info.Server = normalizeServer(got)

	var want string
	if doc.Cluster.ExpectedServer != "" {
		want = doc.Cluster.ExpectedServer
	} else if discovery.OutputsPath != "" {
		want, err = clusterEndpointFromOutputs(discovery.OutputsPath)
		if err != nil {
			return info, err
		}
	}
	if want == "" {
		return info, nil
	}
	want = normalizeServer(want)
	if want != info.Server {
		return info, IdentityGuardRefusal(want, got)
	}
	info.Expected = want
	info.Verified = true
	return info, nil
}

func clusterEndpointFromOutputs(dir string) (string, error) {
	path := filepath.Join(dir, "app-platform-outputs.json")
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("infrastructure outputs: read %s: %w", path, err)
	}
	var flat map[string]any
	if err := json.Unmarshal(b, &flat); err != nil {
		return "", err
	}
	if v, ok := flat["cluster_endpoint"].(string); ok && v != "" {
		return v, nil
	}
	return "", nil
}

func normalizeServer(u string) string {
	u = strings.TrimSpace(u)
	u = strings.TrimSuffix(u, "/")
	return u
}