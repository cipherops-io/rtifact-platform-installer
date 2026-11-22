package cloud

import (
	"context"

	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

// Defaults supplies schema, maturity, and quota behaviour for a cloud.
type Defaults interface {
	DefaultMaturity() string
	CustomerSchemaUsesCommonAllOf() bool
	DefaultQuotaCheckOnValidate() bool
}

// Config supplies cloud-specific merge, validation, and preflight hooks.
type Config interface {
	TFVarSections() []string
	Normalize(merged map[string]any)
	Validate(merged map[string]any) error
	Preflight(ctx context.Context, merged map[string]any) error
}

// Lifecycle covers cloud-specific apply/destroy helpers (DNS, orphan guard).
type Lifecycle interface {
	PrepareLayerVars(ctx context.Context, spec StateSpec, layer stacks.Layer, merged map[string]any) error
	DestroyOrphanGuard(layer stacks.Layer, region, namePrefix string, merged map[string]any) error
	MutateDNSDelegation(ctx context.Context, layer stacks.Layer, region, namePrefix, networkEdgeOutputsPath string, merged map[string]any)
	DelegationPauseIfNeeded(merged map[string]any, configPath, namePrefix string, flat map[string]any) error
}

// Cloud is the full installer surface for one supported cloud.
type Cloud interface {
	Provider
	Defaults
	Config
	Lifecycle
}