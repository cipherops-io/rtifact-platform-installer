package deploy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
)

// CheckIdentity compares kubeconfig server to expected_server or infra outputs.
func CheckIdentity(kcfg kube.Config, doc *Document, discovery Discovery) error {
	got, err := kube.ServerURL(kcfg)
	if err != nil {
		return err
	}
	var want string
	if doc.Cluster.ExpectedServer != "" {
		want = doc.Cluster.ExpectedServer
	} else if discovery.OutputsPath != "" {
		want, err = clusterEndpointFromOutputs(discovery.OutputsPath)
		if err != nil {
			return err
		}
	}
	if want == "" {
		return nil
	}
	if normalizeServer(want) != normalizeServer(got) {
		return IdentityGuardRefusal(want, got)
	}
	return nil
}

func clusterEndpointFromOutputs(dir string) (string, error) {
	path := filepath.Join(dir, "app-platform-outputs.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("infrastructure outputs: read %s: %w", path, err)
	}
	var flat map[string]any
	if err := json.Unmarshal(b, &flat); err != nil {
		return "", err
	}
	if v, ok := flat["cluster_endpoint"].(string); ok && v != "" {
		return v, nil
	}
	return "", fmt.Errorf("infrastructure outputs missing cluster_endpoint in %s", path)
}

func normalizeServer(u string) string {
	u = strings.TrimSpace(u)
	u = strings.TrimSuffix(u, "/")
	return u
}