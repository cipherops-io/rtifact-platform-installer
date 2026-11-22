package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
	"github.com/cipherops-io/rtifact-platform-installer/internal/checklist"
	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud"
	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud/registry"
	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/engine"
	"github.com/cipherops-io/rtifact-platform-installer/internal/outputs"
	"github.com/cipherops-io/rtifact-platform-installer/internal/paths"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
)

type Options struct {
	Action               string
	ConfigPath           string
	Stack                stacks.Layer
	Yes                  bool
	AllowDestroyStateful bool
	PrintChecklist       bool
	PlanOut              string
	DestroyBatch         []stacks.Layer
	IgnoreRuntimeDeps    bool
	DryRunDestroy        bool
	AcceptPreview        bool

	layerDeployed func(stacks.Layer) bool
}

func Run(ctx context.Context, opts Options) error {
	runCtx, cleanup, err := loadRunContext(ctx, opts)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}
	if err := runPreLayers(ctx, opts, runCtx); err != nil {
		return err
	}
	targetLayers, opts := resolveTargetLayers(opts)
	if opts.Action == "destroy" && opts.DryRunDestroy {
		return destroyDryRunReport(opts, runCtx.customer.NamePrefix(), targetLayers, runCtx.probe.deployed)
	}
	if err := runLayers(ctx, opts, runCtx, targetLayers); err != nil {
		return err
	}
	return runPostApply(opts, runCtx, targetLayers)
}

type runContext struct {
	customer    *config.CustomerConfig
	meta        *config.ProfileMeta
	provider    cloud.Cloud
	spec        cloud.StateSpec
	merged      map[string]any
	contentRoot string
	probe       *deployProbe
}

func loadRunContext(ctx context.Context, opts Options) (*runContext, func(), error) {
	if err := config.ValidateCustomerYAML(opts.ConfigPath); err != nil {
		return nil, nil, err
	}
	customer, err := config.LoadCustomer(opts.ConfigPath)
	if err != nil {
		return nil, nil, err
	}
	contentRoot, isTemp, err := assets.ContentRoot()
	if err != nil {
		return nil, nil, fmt.Errorf("locate profiles/infra: %w", err)
	}
	var cleanup func()
	if isTemp {
		cleanup = func() { os.RemoveAll(contentRoot) }
	}
	profilesDir := filepath.Join(contentRoot, "profiles")
	defaults, err := config.LoadProfileDefaults(profilesDir, customer.Profile, customer.Cloud)
	if err != nil {
		return nil, nil, err
	}
	meta, err := config.LoadProfileMeta(profilesDir, customer.Profile)
	if err != nil {
		return nil, nil, err
	}
	merged, err := config.MergeBuildFromPath(defaults, opts.ConfigPath)
	if err != nil {
		return nil, nil, err
	}
	if err := registry.EnsureOperational(customer.Cloud); err != nil {
		return nil, nil, err
	}
	provider, err := registry.Get(customer.Cloud)
	if err != nil {
		return nil, nil, err
	}
	spec := config.CloudStateSpec(customer, meta.StackID)
	probe := newDeployProbe(ctx, contentRoot, customer, meta, provider, spec, merged)
	return &runContext{customer, meta, provider, spec, merged, contentRoot, probe}, cleanup, nil
}

func runPreLayers(ctx context.Context, opts Options, rc *runContext) error {
	opts.layerDeployed = rc.probe.deployed
	if opts.Action != "apply" {
		return nil
	}
	maturity := rc.meta.CloudMaturityFor(rc.customer.Cloud)
	if err := previewCloudGate(maturity, rc.customer.Cloud, opts.AcceptPreview); err != nil {
		return err
	}
	return ensureBackendBeforeApplyAll(ctx, opts, rc.customer, rc.provider, rc.spec, opts.ConfigPath)
}

func resolveTargetLayers(opts Options) ([]stacks.Layer, Options) {
	stack := opts.Stack
	if stack == "" {
		stack = stacks.LayerAll
	}
	targetLayers := stacks.ResolveLayers(stack)
	if opts.Action == "destroy" && stack == stacks.LayerAll {
		targetLayers = stacks.DestroyOrder()
		opts.DestroyBatch = append([]stacks.Layer(nil), targetLayers...)
	}
	return targetLayers, opts
}

func runLayers(ctx context.Context, opts Options, rc *runContext, layers []stacks.Layer) error {
	for _, layer := range layers {
		if err := applyUpstreamGuard(opts, layer, rc.probe.deployed, opts.ConfigPath); err != nil {
			return err
		}
		lr := &layerRun{
			customer:    rc.customer,
			meta:        rc.meta,
			provider:    rc.provider,
			spec:        rc.spec,
			contentRoot: rc.contentRoot,
			merged:      rc.merged,
			layer:       layer,
		}
		step := newLayerStep(ctx, opts, lr)
		if err := engine.Run(ctx, []engine.Step{step}); err != nil {
			return usererr.Wrap(err,
				fmt.Sprintf("OpenTofu %s failed on stack %q.", opts.Action, layer),
				"Read the OpenTofu output above, fix the cause, then re-run with the same --stack.")
		}
	}
	return nil
}

func runPostApply(opts Options, rc *runContext, targetLayers []stacks.Layer) error {
	if opts.Action != "apply" {
		return nil
	}
	if opts.PrintChecklist && rc.meta.ChecklistID != "" && containsLayer(targetLayers, stacks.LayerAppPlatform) {
		if text, e := checklist.Render(rc.meta.ChecklistID); e == nil {
			fmt.Fprintln(os.Stderr, text)
		}
	}
	if !containsLayer(targetLayers, stacks.LayerAppPlatform) {
		return nil
	}
	written, err := outputs.MaterializeFromArtifacts(rc.customer.NamePrefix(), outputs.MaterializeOpts{
		AWSRegion: rc.customer.Account.Region,
	})
	if err != nil {
		return err
	}
	for _, p := range written {
		fmt.Fprintf(os.Stderr, "Wrote %s\n", p)
	}
	return nil
}

func containsLayer(layers []stacks.Layer, want stacks.Layer) bool {
	for _, l := range layers {
		if l == want {
			return true
		}
	}
	return false
}

func layerSpec(spec cloud.StateSpec, layer stacks.Layer) cloud.StateSpec {
	if layer == stacks.LayerStateBackend {
		spec.Local = true
		spec.LocalDir = filepath.Join(paths.ArtifactsDir(spec.NamePrefix), "state")
		_ = os.MkdirAll(spec.LocalDir, 0o755)
	}
	return spec
}

func stackRelDir(profileStackID, cloudName string, layer stacks.Layer) string {
	return filepath.Join("stacks", profileStackID, cloudName, string(layer))
}