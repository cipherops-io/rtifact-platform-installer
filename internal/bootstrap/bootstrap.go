package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud"
	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud/registry"
	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/runner"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

// Run ensures the remote state backend exists. If it is missing, the
// state-backend layer is applied (with local state), then re-verified.
func Run(configPath string, yes bool) error {
	if err := config.ValidateCustomerYAML(configPath); err != nil {
		return err
	}
	customer, err := config.LoadCustomer(configPath)
	if err != nil {
		return err
	}
	if customer.UsesLocalState() {
		fmt.Fprintln(os.Stderr, "bootstrap: state.backend is local; no remote state backend to create")
		return nil
	}
	if err := registry.EnsureOperational(customer.Cloud); err != nil {
		return err
	}
	provider, err := registry.Get(customer.Cloud)
	if err != nil {
		return err
	}
	root, isTemp, err := assets.ContentRoot()
	if err != nil {
		return fmt.Errorf("locate profiles/infra: %w", err)
	}
	if isTemp {
		defer os.RemoveAll(root)
	}
	meta, err := config.LoadProfileMeta(filepath.Join(root, "profiles"), customer.Profile)
	if err != nil {
		return err
	}
	spec := config.CloudStateSpec(customer, meta.StackID)
	ctx := context.Background()

	info, err := provider.EnsureStateBackend(ctx, spec)
	if err == nil {
		fmt.Fprintf(os.Stderr, "bootstrap: state backend %s/%s already exists\n", info.ResourceGroupName, info.StorageAccountName)
		return nil
	}
	if !errors.Is(err, cloud.ErrStateBackendMissing) {
		return err
	}
	fmt.Fprintf(os.Stderr, "bootstrap: %v; applying %s layer\n", err, stacks.LayerStateBackend)

	if err := runner.Run(runner.Options{
		Action:     "apply",
		ConfigPath: configPath,
		Stack:      stacks.LayerStateBackend,
		Yes:        yes,
	}); err != nil {
		return err
	}
	if _, err := provider.EnsureStateBackend(ctx, spec); err != nil {
		return fmt.Errorf("state backend not usable after bootstrap: %w", err)
	}
	return nil
}
