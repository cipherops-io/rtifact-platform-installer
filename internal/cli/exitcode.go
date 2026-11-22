package cli

import (
	"errors"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy"
	"github.com/cipherops-io/rtifact-platform-installer/internal/engine"
)

// ExitCode maps errors to process exit status.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	if errors.Is(err, deploy.ErrInterrupted) {
		return 130
	}
	if errors.Is(err, engine.ErrAborted) {
		return 1
	}
	return 1
}