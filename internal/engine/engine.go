package engine

import "context"

// Step is one ordered phase of installer work.
type Step interface {
	Run(ctx context.Context) error
}

// Options controls interactive gates inside the engine.
type Options struct {
	Yes bool
}

// Change is a pending stack mutation that may require confirmation.
type Change struct {
	Action string
	Stack  string
}

// Run executes steps in order; the first error stops the run.
func Run(ctx context.Context, steps []Step) error {
	for _, s := range steps {
		if err := s.Run(ctx); err != nil {
			return err
		}
	}
	return nil
}

// ConfirmIfNeeded prompts unless opts.Yes is set.
func ConfirmIfNeeded(opts Options, ch Change) error {
	if opts.Yes {
		return nil
	}
	switch ch.Action {
	case "apply":
		return ConfirmApply(ch.Stack)
	case "destroy":
		return ConfirmDestroy(ch.Stack)
	default:
		return nil
	}
}