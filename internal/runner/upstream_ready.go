package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud"
	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

// upstreamReadyForPlan reports whether upstream stacks have been applied enough
// for plan on layer (outputs present in remote/local state).
func upstreamReadyForPlan(
	ctx context.Context,
	customer *config.CustomerConfig,
	meta *config.ProfileMeta,
	provider cloud.Provider,
	spec cloud.StateSpec,
	contentRoot string,
	merged map[string]any,
	layer stacks.Layer,
) (stacks.Layer, bool, error) {
	for _, up := range stacks.Upstream(layer) {
		ub, err := cloud.Backend(provider, layerSpec(spec, up), string(up))
		if err != nil {
			return "", false, err
		}
		if ub.Type == "local" {
			path := ub.Settings["path"]
			if path == "" {
				continue
			}
			if _, err := os.Stat(path); err != nil {
				return up, false, nil
			}
			continue
		}
		has, err := upstreamRemoteOutputsPresent(ctx, customer, meta, contentRoot, provider, spec, merged, up)
		if err != nil {
			return "", false, err
		}
		if !has {
			return up, false, nil
		}
	}
	return "", true, nil
}

func upstreamRemoteOutputsPresent(
	ctx context.Context,
	customer *config.CustomerConfig,
	meta *config.ProfileMeta,
	contentRoot string,
	provider cloud.Provider,
	spec cloud.StateSpec,
	merged map[string]any,
	upstream stacks.Layer,
) (bool, error) {
	tofuPath, mirrorDir, cleanup, err := resolveTofu(contentRoot)
	if err != nil {
		return false, err
	}
	if cleanup != nil {
		defer cleanup()
	}

	work, err := os.MkdirTemp("", "installer-upstream-check-")
	if err != nil {
		return false, err
	}
	defer os.RemoveAll(work)

	infraWork := filepath.Join(work, "infra")
	if err := copyDir(filepath.Join(contentRoot, "infra"), infraWork); err != nil {
		return false, err
	}
	stackRel := stackRelDir(meta.StackID, customer.Cloud, upstream)
	stackWork := filepath.Join(infraWork, stackRel)

	cl, ok := provider.(cloud.Lifecycle)
	if !ok {
		return false, fmt.Errorf("cloud %q: missing lifecycle hooks", customer.Cloud)
	}
	layerMerged, err := prepareLayerMerged(ctx, provider, cl, spec, upstream, merged, upstreamSpecPerLayer)
	if err != nil {
		return false, err
	}
	if err := config.WriteAutoTofuVarsJSON(stackWork, layerMerged); err != nil {
		return false, err
	}

	backend, err := cloud.Backend(provider, layerSpec(spec, upstream), string(upstream))
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(filepath.Join(stackWork, "backend.tofu"), []byte(backend.HCL()), 0o644); err != nil {
		return false, err
	}

	env := append(os.Environ(), "TF_IN_AUTOMATION=1")
	if mirrorDir != "" {
		rcPath := filepath.Join(work, "tofu.rc")
		if err := writeProviderRC(rcPath, mirrorDir); err != nil {
			return false, err
		}
		env = append(env, "TF_CLI_CONFIG_FILE="+rcPath)
	} else {
		pc, err := pluginCacheDir(contentRoot)
		if err != nil {
			return false, err
		}
		env = append(env, "TF_PLUGIN_CACHE_DIR="+pc)
	}

	initArgs := []string{"init", "-input=false", "-reconfigure"}
	if _, err := os.Stat(filepath.Join(stackWork, ".terraform.lock.hcl")); err == nil {
		initArgs = append(initArgs, "-lockfile=readonly")
	}
	if err := runTofuInit(tofuPath, stackWork, env, initArgs); err != nil {
		return false, nil
	}

	c := exec.Command(tofuPath, "output", "-json")
	c.Dir = stackWork
	c.Env = env
	out, err := c.Output()
	if err != nil {
		return false, nil
	}
	var parsed map[string]json.RawMessage
	if json.Unmarshal(out, &parsed) != nil || len(parsed) == 0 {
		return false, nil
	}
	return true, nil
}