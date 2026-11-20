// Package registry resolves a cloud.Provider from the `cloud:` config value.
package registry

import (
	"fmt"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud"
	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud/aws"
	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud/azure"
	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud/gcp"
	"github.com/cipherops-io/rtifact-platform-installer/internal/version"
)

// Supported lists clouds compiled into this binary.
var Supported = version.PackagedCloudsList()

func init() {
	if len(Supported) == 0 {
		Supported = []string{azure.CloudName, aws.CloudName, gcp.CloudName}
	}
}

// Get returns the provider for name.
func Get(name string) (cloud.Provider, error) {
	if !version.CloudPackaged(name) {
		return nil, fmt.Errorf("cloud %q is not in this binary (packaged: %v)", name, Supported)
	}
	switch name {
	case azure.CloudName:
		return azure.New(), nil
	case aws.CloudName:
		return aws.New(), nil
	case gcp.CloudName:
		return gcp.New(), nil
	default:
		return nil, fmt.Errorf("unsupported cloud %q (supported: %v)", name, Supported)
	}
}

// EnsureOperational returns an error if cloud is not operational in this build.
func EnsureOperational(name string) error {
	if !version.CloudPackaged(name) {
		return fmt.Errorf("cloud %q is not packaged in this binary", name)
	}
	for _, c := range version.OperationalClouds() {
		if c == name {
			return nil
		}
	}
	return fmt.Errorf("cloud %q is not operational (available: %s)", name, strings.Join(version.OperationalClouds(), ", "))
}
