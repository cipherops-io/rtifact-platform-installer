package gcp

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud"
)

type CommandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

type Provider struct {
	cloud.NoopLifecycle
	run      CommandRunner
	lookPath func(string) (string, error)
}

var _ cloud.Cloud = (*Provider)(nil)

func New() *Provider {
	return &Provider{run: execRun, lookPath: exec.LookPath}
}

func execRun(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).Output()
}

func (*Provider) Name() string { return CloudName }

func stateBucketName(namePrefix string) string {
	base := strings.ToLower(strings.ReplaceAll(namePrefix, "_", "-"))
	if len(base) > 50 {
		base = base[:50]
	}
	return base + StateBucketSuffix
}

func stateBackendNames(spec cloud.StateSpec) string {
	bucket := stateBucketName(spec.NamePrefix)
	if o := spec.Override; o != nil && o.GCSBucket != "" {
		bucket = o.GCSBucket
	}
	return bucket
}

func (*Provider) Backend(spec cloud.StateSpec, layer string) (*cloud.BackendConfig, error) {
	if spec.Local {
		return cloud.LocalBackend(spec, layer)
	}
	if spec.Account.ProjectID == "" {
		return nil, errors.New("gcp backend requires account.project_id")
	}
	bucket := stateBackendNames(spec)
	prefix := cloud.StateKey(spec, layer)
	settings := map[string]string{
		"bucket":  bucket,
		"prefix":  strings.TrimSuffix(prefix, ".tfstate"),
		"project": spec.Account.ProjectID,
	}
	return &cloud.BackendConfig{Type: "gcs", Settings: settings}, nil
}

type StateBackendMissingError struct {
	Resource string
}

func (e *StateBackendMissingError) Error() string {
	return fmt.Sprintf("%v: %s (run `rtifact bootstrap`)", cloud.ErrStateBackendMissing, e.Resource)
}

func (e *StateBackendMissingError) Unwrap() error { return cloud.ErrStateBackendMissing }

func (p *Provider) EnsureStateBackend(ctx context.Context, spec cloud.StateSpec) (*cloud.StateBackendInfo, error) {
	if spec.Account.ProjectID == "" {
		return nil, errors.New("gcp EnsureStateBackend requires account.project_id")
	}
	if _, err := p.lookPath(CLI); err != nil {
		return nil, fmt.Errorf("%s CLI not found: %w", CLI, err)
	}
	bucket := stateBackendNames(spec)
	if _, err := p.run(ctx, CLI, "storage", "buckets", "describe", "gs://"+bucket, "--project", spec.Account.ProjectID); err != nil {
		return nil, &StateBackendMissingError{Resource: "gcs bucket " + bucket}
	}
	return &cloud.StateBackendInfo{
		ResourceGroupName:  spec.Account.ProjectID,
		GCSBucket:          bucket,
		StorageAccountName: bucket,
	}, nil
}
