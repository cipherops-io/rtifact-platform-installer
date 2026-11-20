package testutil

import (
	"os"
	"path/filepath"
	"runtime"
)

// InfraRoot is the OpenTofu infra tree (stacks + modules) used by stack layers.
// Override with TERRATEST_INFRA_ROOT (absolute or relative to cwd).
func InfraRoot() string {
	if v := os.Getenv("TERRATEST_INFRA_ROOT"); v != "" {
		if filepath.IsAbs(v) {
			return v
		}
		if cwd, err := os.Getwd(); err == nil {
			return filepath.Join(cwd, v)
		}
		return v
	}
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	// .../rtifact-platform-installer/tests/stacks/terratest/testutil/infra_root.go
	terratestDir := filepath.Dir(filepath.Dir(thisFile))
	installerRoot := filepath.Join(terratestDir, "..", "..", "..")
	return filepath.Join(installerRoot, "infra")
}

// StackLayerDir returns infra/stacks/aios-infra/<cloud>/<layer>.
func StackLayerDir(cloud, layer string) string {
	return filepath.Join(InfraRoot(), "stacks", "aios-infra", cloud, layer)
}

func CoreNetworkStateFixture() string {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	terratestDir := filepath.Dir(filepath.Dir(thisFile))
	return filepath.Join(terratestDir, "testdata", "core-network.tfstate")
}

func VarFileForCloud(cloud string) string {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	terratestDir := filepath.Dir(filepath.Dir(thisFile))
	return filepath.Join(terratestDir, "examples", cloud+".auto.tfvars.json")
}