package version

import (
	"fmt"
	"runtime/debug"

	_ "helm.sh/helm/v3/pkg/chart" // keeps Helm SDK linked in release binaries
)

// DeployEngineLine reports the bundled Helm module version from build metadata.
func DeployEngineLine() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "deploy-engine: Helm (unknown)"
	}
	for _, m := range info.Deps {
		if m.Path == "helm.sh/helm/v3" {
			v := m.Version
			if v == "" {
				v = "(devel)"
			}
			return fmt.Sprintf("deploy-engine: Helm %s", v)
			}
	}
	return "deploy-engine: Helm (not linked)"
}