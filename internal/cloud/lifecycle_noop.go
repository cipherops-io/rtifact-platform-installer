package cloud

import (
	"context"

	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

// NoopLifecycle provides default no-op lifecycle hook methods for clouds without
// extra destroy/DNS behaviour.
type NoopLifecycle struct{}

func (NoopLifecycle) PrepareLayerVars(_ context.Context, _ StateSpec, _ stacks.Layer, _ map[string]any) error {
	return nil
}

func (NoopLifecycle) DestroyOrphanGuard(_ stacks.Layer, _, _ string, _ map[string]any) error {
	return nil
}

func (NoopLifecycle) MutateDNSDelegation(_ context.Context, _ stacks.Layer, _, _, _ string, _ map[string]any) {
}

func (NoopLifecycle) DelegationPauseIfNeeded(_ map[string]any, _, _ string, _ map[string]any) error {
	return nil
}