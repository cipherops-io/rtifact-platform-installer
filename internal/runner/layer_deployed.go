package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud"
	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

// deployProbe caches whether a layer has resources in OpenTofu state (one CLI run).
type deployProbe struct {
	ctx         context.Context
	contentRoot string
	customer    *config.CustomerConfig
	meta        *config.ProfileMeta
	provider    cloud.Provider
	spec        cloud.StateSpec
	merged      map[string]any

	mu     sync.Mutex
	cache  map[stacks.Layer]bool
	marker func(string, stacks.Layer) bool
}

func newDeployProbe(
	ctx context.Context,
	contentRoot string,
	customer *config.CustomerConfig,
	meta *config.ProfileMeta,
	provider cloud.Provider,
	spec cloud.StateSpec,
	merged map[string]any,
) *deployProbe {
	return &deployProbe{
		ctx:         ctx,
		contentRoot: contentRoot,
		customer:    customer,
		meta:        meta,
		provider:    provider,
		spec:        spec,
		merged:      merged,
		cache:       make(map[stacks.Layer]bool),
		marker:      layerDeploymentMarker,
	}
}

func (p *deployProbe) deployed(layer stacks.Layer) bool {
	p.mu.Lock()
	if v, ok := p.cache[layer]; ok {
		p.mu.Unlock()
		return v
	}
	p.mu.Unlock()

	ok, err := p.layerHasState(layer)
	if err != nil {
		ok = p.marker(p.customer.NamePrefix(), layer)
	}
	p.mu.Lock()
	p.cache[layer] = ok
	p.mu.Unlock()
	return ok
}

func (p *deployProbe) layerHasState(layer stacks.Layer) (bool, error) {
	if layer == stacks.LayerStateBackend {
		path := filepath.Join(layerSpec(p.spec, layer).LocalDir, "state-backend.tfstate")
		if st, err := os.Stat(path); err == nil && st.Size() > 0 {
			return true, nil
		}
		return false, nil
	}

	tofuPath, mirrorDir, cleanup, err := resolveTofu(p.contentRoot)
	if err != nil {
		return false, err
	}
	if cleanup != nil {
		defer cleanup()
	}

	work, err := os.MkdirTemp("", "installer-state-probe-")
	if err != nil {
		return false, err
	}
	defer os.RemoveAll(work)

	infraWork := filepath.Join(work, "infra")
	if err := copyDir(filepath.Join(p.contentRoot, "infra"), infraWork); err != nil {
		return false, err
	}
	stackRel := stackRelDir(p.meta.StackID, p.customer.Cloud, layer)
	stackWork := filepath.Join(infraWork, stackRel)

	cl, ok := p.provider.(cloud.Lifecycle)
	if !ok {
		return false, fmt.Errorf("cloud %q: missing lifecycle hooks", p.customer.Cloud)
	}
	layerMerged, err := prepareLayerMerged(p.ctx, p.provider, cl, p.spec, layer, p.merged, upstreamSpecPerLayer)
	if err != nil {
		return false, err
	}
	if err := config.WriteAutoTofuVarsJSON(stackWork, layerMerged); err != nil {
		return false, err
	}
	backend, err := cloud.Backend(p.provider, layerSpec(p.spec, layer), string(layer))
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
		pc, err := pluginCacheDir(p.contentRoot)
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
		return false, err
	}

	c := exec.Command(tofuPath, "state", "list")
	c.Dir = stackWork
	c.Env = env
	var stderr bytes.Buffer
	c.Stderr = &stderr
	out, err := c.Output()
	if err != nil {
		return false, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, ln := range lines {
		if strings.TrimSpace(ln) != "" {
			return true, nil
		}
	}
	return false, nil
}