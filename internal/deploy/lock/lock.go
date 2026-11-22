package lock

import (
	"context"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
)

// Handle is an exclusive deploy lock (Lease in rtifact-system).
type Handle struct {
	Holder string
}

// Acquire takes the deployment lock. C2: no-op; C3: coordination.Lease.
func Acquire(ctx context.Context, kcfg kube.Config, deployment, holder string, force bool) (*Handle, error) {
	_ = ctx
	_ = kcfg
	_ = deployment
	_ = holder
	_ = force
	return &Handle{Holder: holder}, nil
}

// Release drops the lock.
func (h *Handle) Release(ctx context.Context, kcfg kube.Config, deployment string) error {
	_ = ctx
	_ = kcfg
	_ = deployment
	return nil
}