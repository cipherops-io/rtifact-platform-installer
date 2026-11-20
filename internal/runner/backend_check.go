package runner

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud"
	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
)

func ensureBackendBeforeApplyAll(opts Options, customer *config.CustomerConfig, provider cloud.Provider, spec cloud.StateSpec, configPath string) error {
	if opts.Action != "apply" || opts.Stack != stacks.LayerAll {
		return nil
	}
	if customer.UsesLocalState() {
		return nil
	}
	if customer.Backend != nil {
		_, err := provider.EnsureStateBackend(context.Background(), spec)
		if errors.Is(err, cloud.ErrStateBackendMissing) {
			return usererr.WithHint(
				"Customer backend block points at resources that do not exist.",
				"Create the backend in your subscription or fix backend: in your config.",
				err.Error(),
			)
		}
		return err
	}
	_, err := provider.EnsureStateBackend(context.Background(), spec)
	if err == nil {
		return nil
	}
	if !errors.Is(err, cloud.ErrStateBackendMissing) {
		return err
	}
	fmt.Fprintf(os.Stderr, "apply --stack all: %v; run bootstrap first\n", err)
	return usererr.WithHint(
		"Remote state backend is missing.",
		fmt.Sprintf("Run: rtifact bootstrap %s", configPath),
		err.Error(),
	)
}