package ocireg

import (
	"fmt"
	"strings"

	"helm.sh/helm/v4/pkg/registry"
)

// ChartDigest resolves the OCI artifact digest for chart at version (no pull).
func ChartDigest(client *registry.Client, reg Registry, chart, version string) (string, error) {
	if client == nil || version == "" {
		return "", nil
	}
	ref, err := ChartRef(reg, chart)
	if err != nil {
		return "", err
	}
	path := strings.TrimPrefix(ref, "oci://")
	if path == "" {
		return "", fmt.Errorf("chart ref %q: missing oci host", ref)
	}
	desc, err := client.Resolve(path + ":" + version)
	if err != nil {
		return "", err
	}
	if desc.Digest == "" {
		return "", nil
	}
	return desc.Digest.String(), nil
}