package runner

import (
	"context"

	"github.com/cipherops-io/rtifact-platform-installer/internal/engine"
)

type layerStep struct {
	lr *layerRun
}

func newLayerStep(ctx context.Context, opts Options, lr *layerRun) *layerStep {
	lr.ctx = ctx
	lr.opts = opts
	return &layerStep{lr: lr}
}

func (s *layerStep) Run(ctx context.Context) error {
	defer func() {
		if s.lr.tofuCleanup != nil {
			s.lr.tofuCleanup()
			s.lr.tofuCleanup = nil
		}
	}()
	return engine.Run(ctx, []engine.Step{
		&layerGuardStep{lr: s.lr},
		&layerWorkspaceStep{lr: s.lr},
		&layerTofuStep{lr: s.lr},
	})
}