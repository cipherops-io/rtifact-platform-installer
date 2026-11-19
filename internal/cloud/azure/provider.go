package azure

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud"
)

// CommandRunner executes an external command and returns its stdout.
type CommandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

// Provider implements cloud.Provider for Azure.
type Provider struct {
	run CommandRunner
	// lookPath resolves the az executable (swapped in tests).
	lookPath func(string) (string, error)
}

var _ cloud.Provider = (*Provider)(nil)

// New returns an Azure provider that shells out to the Azure CLI.
func New() *Provider {
	return &Provider{run: execRun, lookPath: exec.LookPath}
}

// NewWithRunner returns a provider using a custom command runner (tests).
func NewWithRunner(run CommandRunner) *Provider {
	return &Provider{run: run, lookPath: func(s string) (string, error) { return s, nil }}
}

func execRun(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).Output()
}

// Name implements cloud.Provider.
func (*Provider) Name() string { return CloudName }

// StateStorageAccountName is the storage account created by the state-backend layer.
// It must match infra/modules/state-backend.
func StateStorageAccountName(namePrefix string) string {
	base := strings.ReplaceAll(namePrefix, "-", "")
	maxBase := StorageAccountNameMaxLen - len(StateStorageAccountSuffix)
	if len(base) > maxBase {
		base = base[:maxBase]
	}
	return base + StateStorageAccountSuffix
}

func stateBackendNames(spec cloud.StateSpec) cloud.StateBackendInfo {
	info := cloud.StateBackendInfo{
		ResourceGroupName:  SharedResourceGroupName(spec),
		StorageAccountName: StateStorageAccountName(spec.NamePrefix),
		ContainerName:      StateContainerName,
	}
	if o := spec.Override; o != nil {
		if o.ResourceGroupName != "" {
			info.ResourceGroupName = o.ResourceGroupName
		}
		if o.StorageAccountName != "" {
			info.StorageAccountName = o.StorageAccountName
		}
		if o.ContainerName != "" {
			info.ContainerName = o.ContainerName
		}
	}
	return info
}

// Backend implements cloud.Provider.
func (*Provider) Backend(spec cloud.StateSpec, layer string) (*cloud.BackendConfig, error) {
	if spec.Local {
		return cloud.LocalBackend(spec, layer)
	}
	if spec.Account.SubscriptionID == "" {
		return nil, errors.New("azure backend requires account.subscription_id")
	}
	info := stateBackendNames(spec)
	settings := map[string]string{
		"resource_group_name":  info.ResourceGroupName,
		"storage_account_name": info.StorageAccountName,
		"container_name":       info.ContainerName,
		"key":                  cloud.StateKey(spec, layer),
		"subscription_id":      spec.Account.SubscriptionID,
		"use_azuread_auth":     "true",
	}
	if spec.Account.TenantID != "" {
		settings["tenant_id"] = spec.Account.TenantID
	}
	if spec.Account.Environment != "" {
		settings["environment"] = spec.Account.Environment
	}
	if spec.Account.UseMSI {
		settings["use_msi"] = "true"
	}
	return &cloud.BackendConfig{Type: "azurerm", Settings: settings}, nil
}

// StateBackendMissingError names the Azure resource that was not found.
type StateBackendMissingError struct {
	Resource string
}

func (e *StateBackendMissingError) Error() string {
	return fmt.Sprintf("%v: %s (run `installer bootstrap`)", cloud.ErrStateBackendMissing, e.Resource)
}

func (e *StateBackendMissingError) Unwrap() error { return cloud.ErrStateBackendMissing }

// EnsureStateBackend implements cloud.Provider.
//
// Only the air-gapped (Azure CLI) path is implemented: it verifies the
// subscription, resource group, storage account and container with `az`.
// It never creates anything; creation is done by the state-backend layer.
func (p *Provider) EnsureStateBackend(ctx context.Context, spec cloud.StateSpec) (*cloud.StateBackendInfo, error) {
	if !spec.AirGapped {
		return nil, fmt.Errorf("azure EnsureStateBackend with state.air_gapped=false: %w", cloud.ErrNotImplemented)
	}
	if spec.Account.SubscriptionID == "" {
		return nil, errors.New("azure EnsureStateBackend requires account.subscription_id")
	}
	if _, err := p.lookPath(AzureCLI); err != nil {
		return nil, fmt.Errorf("%s CLI not found on PATH (required for air-gapped state backend checks): %w", AzureCLI, err)
	}
	info := stateBackendNames(spec)
	sub := spec.Account.SubscriptionID

	out, err := p.az(ctx, "account", "show", "--subscription", sub, "--query", "id", "-o", "tsv")
	if err != nil {
		return nil, fmt.Errorf("az account show for subscription %s: %w", sub, err)
	}
	if !strings.EqualFold(strings.TrimSpace(out), sub) {
		return nil, fmt.Errorf("az is not logged in to subscription %s", sub)
	}

	out, err = p.az(ctx, "group", "exists", "--name", info.ResourceGroupName, "--subscription", sub)
	if err != nil {
		return nil, fmt.Errorf("az group exists %s: %w", info.ResourceGroupName, err)
	}
	if strings.TrimSpace(out) != "true" {
		return nil, &StateBackendMissingError{Resource: "resource group " + info.ResourceGroupName}
	}

	if _, err := p.az(ctx, "storage", "account", "show",
		"--name", info.StorageAccountName, "--resource-group", info.ResourceGroupName,
		"--subscription", sub, "--query", "name", "-o", "tsv"); err != nil {
		return nil, &StateBackendMissingError{Resource: "storage account " + info.StorageAccountName}
	}

	out, err = p.az(ctx, "storage", "container", "exists",
		"--name", info.ContainerName, "--account-name", info.StorageAccountName,
		"--auth-mode", "login", "--subscription", sub, "--query", "exists", "-o", "tsv")
	if err != nil {
		return nil, fmt.Errorf("az storage container exists %s: %w", info.ContainerName, err)
	}
	if strings.TrimSpace(out) != "true" {
		return nil, &StateBackendMissingError{Resource: "container " + info.ContainerName}
	}
	return &info, nil
}

func (p *Provider) az(ctx context.Context, args ...string) (string, error) {
	out, err := p.run(ctx, AzureCLI, args...)
	return string(out), err
}
