package deploy

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
)

// ErrInterrupted is returned when deploy is cancelled by SIGINT/SIGTERM (CLI maps to exit 130).
var ErrInterrupted = errors.New("deploy interrupted")

// SignalContext wraps parent with SIGINT/SIGTERM cancellation (C1).
func SignalContext(parent context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
}

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