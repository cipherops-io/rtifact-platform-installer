// Package cloud defines the provider-neutral surface the installer uses to talk
// to a target cloud (state backend discovery and backend configuration).
package cloud

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// ErrNotImplemented is returned for code paths that are not supported yet.
var ErrNotImplemented = errors.New("not implemented")

// ErrStateBackendMissing is returned by EnsureStateBackend when the remote state
// backend does not exist yet (run `installer bootstrap`).
var ErrStateBackendMissing = errors.New("state backend does not exist")

// Account identifies the cloud account/subscription the installer targets.
type Account struct {
	SubscriptionID string
	TenantID       string
	AccountID      string // AWS
	ProjectID      string // GCP
	Region         string
	Environment    string
	UseMSI         bool
}

// BackendOverride lets a customer point at a pre-created state backend.
// Key, when set, is used as a prefix; the layer name is always appended so that
// layers never share a state file.
type BackendOverride struct {
	ResourceGroupName  string
	StorageAccountName string
	ContainerName      string
	Key                string
	S3Bucket           string
	DynamoDBTable      string
	S3Region           string
	GCSBucket          string
	GCSPrefix          string
}

// StateSpec describes where state for a workload lives.
type StateSpec struct {
	Profile    string
	NamePrefix string
	Account    Account
	// Local selects the local backend instead of the cloud remote backend.
	Local bool
	// LocalDir is the directory holding local state files (Local only).
	LocalDir string
	// AirGapped means no cloud SDKs or control-plane reachability beyond the CLI.
	AirGapped bool
	Override  *BackendOverride
}

// StateBackendInfo describes an existing remote state backend.
type StateBackendInfo struct {
	ResourceGroupName  string
	StorageAccountName string
	ContainerName      string
	S3Bucket           string
	DynamoDBTable      string
	GCSBucket          string
}

// BackendConfig is an OpenTofu backend block ("azurerm", "local", ...).
type BackendConfig struct {
	Type     string
	Settings map[string]string
}

// Provider is implemented once per supported cloud.
type Provider interface {
	// Name is the value of the `cloud:` field in customer config.
	Name() string
	// EnsureStateBackend verifies the remote state backend exists.
	EnsureStateBackend(ctx context.Context, spec StateSpec) (*StateBackendInfo, error)
	// Backend returns the backend configuration for one layer.
	Backend(spec StateSpec, layer string) (*BackendConfig, error)
}

// Backend returns the backend configuration for layer using p.
func Backend(p Provider, spec StateSpec, layer string) (*BackendConfig, error) {
	if p == nil {
		return nil, errors.New("cloud provider is nil")
	}
	return p.Backend(spec, layer)
}

// LocalBackend returns a local backend for layer under spec.LocalDir.
func LocalBackend(spec StateSpec, layer string) (*BackendConfig, error) {
	if spec.LocalDir == "" {
		return nil, errors.New("local state requires a state directory")
	}
	return &BackendConfig{
		Type:     "local",
		Settings: map[string]string{"path": filepath.Join(spec.LocalDir, layer+".tfstate")},
	}, nil
}

// StateKey is the remote state object name for a layer.
func StateKey(spec StateSpec, layer string) string {
	if spec.Override != nil && spec.Override.Key != "" {
		return strings.TrimSuffix(spec.Override.Key, ".tfstate") + "/" + layer + ".tfstate"
	}
	return fmt.Sprintf("%s/%s/%s.tfstate", spec.Profile, spec.NamePrefix, layer)
}

// HCL renders the backend as a `terraform { backend ... }` block.
func (b *BackendConfig) HCL() string {
	keys := make([]string, 0, len(b.Settings))
	for k := range b.Settings {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	fmt.Fprintf(&sb, "terraform {\n  backend %q {\n", b.Type)
	for _, k := range keys {
		v := b.Settings[k]
		if v == "true" || v == "false" {
			fmt.Fprintf(&sb, "    %s = %s\n", k, v)
		} else {
			fmt.Fprintf(&sb, "    %s = %q\n", k, v)
		}
	}
	sb.WriteString("  }\n}\n")
	return sb.String()
}

// RemoteState returns the value for the `upstream_state` stack variable entry:
// {backend, config} consumed by data "terraform_remote_state".
func (b *BackendConfig) RemoteState() map[string]any {
	cfg := make(map[string]string, len(b.Settings))
	for k, v := range b.Settings {
		cfg[k] = v
	}
	return map[string]any{"backend": b.Type, "config": cfg}
}
