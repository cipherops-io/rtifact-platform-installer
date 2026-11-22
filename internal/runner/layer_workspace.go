package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud"
	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/paths"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

type layerWorkspaceStep struct {
	lr *layerRun
}

func (s *layerWorkspaceStep) Run(ctx context.Context) error {
	lr := s.lr
	if lr.planSkipped {
		return nil
	}
	stackRel := stackRelDir(lr.meta.StackID, lr.customer.Cloud, lr.layer)
	work, err := os.MkdirTemp("", "installer-tofu-")
	if err != nil {
		return err
	}
	lr.workDir = work
	lr.tofuCleanup = func() { os.RemoveAll(work) }

	infraWork := filepath.Join(work, "infra")
	if err := copyDir(filepath.Join(lr.contentRoot, "infra"), infraWork); err != nil {
		return err
	}
	lr.stackWork = filepath.Join(infraWork, stackRel)
	if err := writeLayerVars(ctx, lr); err != nil {
		return err
	}
	return setupTofuRunner(lr, work)
}

func writeLayerVars(ctx context.Context, lr *layerRun) error {
	layerMerged, err := prepareLayerMerged(ctx, lr.provider, lr.provider, lr.spec, lr.layer, lr.merged, upstreamSpecFlat)
	if err != nil {
		return err
	}
	edgeOut := filepath.Join(paths.ArtifactsDir(lr.customer.NamePrefix()), string(stacks.LayerNetworkEdge)+"-outputs.json")
	lr.provider.MutateDNSDelegation(ctx, lr.layer, lr.customer.Account.Region, lr.customer.NamePrefix(), edgeOut, layerMerged)
	if err := config.WriteAutoTofuVarsJSON(lr.stackWork, layerMerged); err != nil {
		return err
	}
	_ = config.WriteMergedConfigArtifact(lr.customer.NamePrefix(), string(lr.layer), layerMerged)

	backend, err := cloud.Backend(lr.provider, layerSpec(lr.spec, lr.layer), string(lr.layer))
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(lr.stackWork, "backend.tofu"), []byte(backend.HCL()), 0o644)
}

func setupTofuRunner(lr *layerRun, work string) error {
	tofuPath, mirrorDir, cleanup, err := resolveTofu(lr.contentRoot)
	if err != nil {
		return err
	}
	lr.tofuPath = tofuPath
	if cleanup != nil {
		prev := lr.tofuCleanup
		lr.tofuCleanup = func() { cleanup(); prev() }
	}
	env, err := tofuEnv(lr.contentRoot, work, mirrorDir)
	if err != nil {
		return err
	}
	lr.tofuEnv = env
	lr.tofu = newTofuRunner(tofuPath, lr.stackWork, env)
	varFile := filepath.Join(lr.stackWork, config.InstallerVarFile)
	lr.varFileArg = "-var-file=" + varFile
	return nil
}

func tofuEnv(contentRoot, work, mirrorDir string) ([]string, error) {
	env := append(os.Environ(), "TF_IN_AUTOMATION=1")
	if mirrorDir != "" {
		rcPath := filepath.Join(work, "tofu.rc")
		if err := writeProviderRC(rcPath, mirrorDir); err != nil {
			return nil, err
		}
		env = append(env, "TF_CLI_CONFIG_FILE="+rcPath)
		return env, nil
	}
	pc, err := pluginCacheDir(contentRoot)
	if err != nil {
		return nil, fmt.Errorf("plugin cache: %w", err)
	}
	env = append(env, "TF_PLUGIN_CACHE_DIR="+pc)
	return env, nil
}