package registry

import (
	"context"
	"fmt"

	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

// LifecycleHooks covers cloud-specific apply/destroy helpers (DNS, orphan guard).
type LifecycleHooks interface {
	DestroyOrphanGuard(layer stacks.Layer, region, namePrefix string, merged map[string]any) error
	MutateDNSDelegation(ctx context.Context, layer stacks.Layer, region, namePrefix, networkEdgeOutputsPath string, merged map[string]any)
	DelegationPauseIfNeeded(merged map[string]any, configPath, namePrefix string, flat map[string]any) error
}

// LifecycleHooksFor returns lifecycle hooks for name.
func LifecycleHooksFor(name string) (LifecycleHooks, error) {
	p, err := Get(name)
	if err != nil {
		return nil, err
	}
	h, ok := p.(LifecycleHooks)
	if !ok {
		return nil, fmt.Errorf("cloud %q: lifecycle hooks not implemented", name)
	}
	return h, nil
}