package deploy

import (
	"context"
	"errors"
)

// ErrInterrupted is returned when apply is cancelled by SIGINT/SIGTERM (CLI maps to exit 130).
var ErrInterrupted = errors.New("deploy apply interrupted")

// AsInterrupted maps context cancellation during apply to ErrInterrupted.
func AsInterrupted(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return errors.Join(ErrInterrupted, err)
	}
	return err
}