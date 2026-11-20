package aws

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
	run      CommandRunner
	lookPath func(string) (string, error)
}

var _ cloud.Provider = (*Provider)(nil)

func New() *Provider {
	return &Provider{run: execRun, lookPath: exec.LookPath}
}

func execRun(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).Output()
}

func (*Provider) Name() string { return CloudName }

func stateBucketName(namePrefix string) string {
	base := strings.ToLower(strings.ReplaceAll(namePrefix, "_", "-"))
	if len(base) > 45 {
		base = base[:45]
	}
	return base + StateBucketSuffix
}

func stateLockTableName(namePrefix string) string {
	base := strings.ReplaceAll(namePrefix, "-", "_")
	if len(base) > 200 {
		base = base[:200]
	}
	return base + StateLockTableSuffix
}

func stateBackendNames(spec cloud.StateSpec) (bucket, table string) {
	bucket = stateBucketName(spec.NamePrefix)
	table = stateLockTableName(spec.NamePrefix)
	if o := spec.Override; o != nil {
		if o.S3Bucket != "" {
			bucket = o.S3Bucket
		}
		if o.DynamoDBTable != "" {
			table = o.DynamoDBTable
		}
	}
	return bucket, table
}

func (*Provider) Backend(spec cloud.StateSpec, layer string) (*cloud.BackendConfig, error) {
	if spec.Local {
		return cloud.LocalBackend(spec, layer)
	}
	if spec.Account.Region == "" {
		return nil, errors.New("aws backend requires account.region")
	}
	bucket, table := stateBackendNames(spec)
	settings := map[string]string{
		"bucket":         bucket,
		"key":            cloud.StateKey(spec, layer),
		"region":         spec.Account.Region,
		"dynamodb_table": table,
		"encrypt":        "true",
	}
	return &cloud.BackendConfig{Type: "s3", Settings: settings}, nil
}

type StateBackendMissingError struct {
	Resource string
}

func (e *StateBackendMissingError) Error() string {
	return fmt.Sprintf("%v: %s (run `rtifact bootstrap`)", cloud.ErrStateBackendMissing, e.Resource)
}

func (e *StateBackendMissingError) Unwrap() error { return cloud.ErrStateBackendMissing }

func (p *Provider) EnsureStateBackend(ctx context.Context, spec cloud.StateSpec) (*cloud.StateBackendInfo, error) {
	if spec.Account.AccountID == "" {
		return nil, errors.New("aws EnsureStateBackend requires account.account_id")
	}
	if _, err := p.lookPath(CLI); err != nil {
		return nil, fmt.Errorf("%s CLI not found: %w", CLI, err)
	}
	bucket, table := stateBackendNames(spec)
	region := spec.Account.Region

	if _, err := p.run(ctx, CLI, "s3api", "head-bucket", "--bucket", bucket); err != nil {
		return nil, &StateBackendMissingError{Resource: "s3 bucket " + bucket}
	}
	if _, err := p.run(ctx, CLI, "dynamodb", "describe-table", "--table-name", table, "--region", region); err != nil {
		return nil, &StateBackendMissingError{Resource: "dynamodb table " + table}
	}
	return &cloud.StateBackendInfo{
		ResourceGroupName:  spec.NamePrefix + "-tfstate",
		S3Bucket:           bucket,
		DynamoDBTable:      table,
		StorageAccountName: bucket,
		ContainerName:      table,
	}, nil
}
