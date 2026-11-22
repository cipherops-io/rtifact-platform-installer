package runner

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
	"github.com/cipherops-io/rtifact-platform-installer/internal/checklist"
	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud"
	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud/registry"
	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/maputil"
	"github.com/cipherops-io/rtifact-platform-installer/internal/paths"
	"github.com/cipherops-io/rtifact-platform-installer/internal/guard"
	"github.com/cipherops-io/rtifact-platform-installer/internal/outputs"
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
	// DestroyBatch is the full layer list for destroy --stack all (dependency guard).
	DestroyBatch []stacks.Layer
	IgnoreRuntimeDeps    bool
	DryRunDestroy        bool
	AcceptPreview        bool

	layerDeployed func(stacks.Layer) bool
}

func Run(opts Options) error {
	if err := config.ValidateCustomerYAML(opts.ConfigPath); err != nil {
		return err
	}
	customer, err := config.LoadCustomer(opts.ConfigPath)
	if err != nil {
		return err
	}
	contentRoot, isTemp, err := assets.ContentRoot()
	if err != nil {
		return fmt.Errorf("locate profiles/infra: %w", err)
	}
	if isTemp {
		defer os.RemoveAll(contentRoot)
	}

	profilesDir := filepath.Join(contentRoot, "profiles")
	defaults, err := config.LoadProfileDefaults(profilesDir, customer.Profile, customer.Cloud)
	if err != nil {
		return err
	}
	meta, err := config.LoadProfileMeta(profilesDir, customer.Profile)
	if err != nil {
		return err
	}
	merged, err := config.MergeBuildFromPath(defaults, opts.ConfigPath)
	if err != nil {
		return err
	}

	if err := registry.EnsureOperational(customer.Cloud); err != nil {
		return err
	}
	provider, err := registry.Get(customer.Cloud)
	if err != nil {
		return err
	}
	spec := config.CloudStateSpec(customer, meta.StackID)
	probe := newDeployProbe(contentRoot, customer, meta, provider, spec, merged)
	opts.layerDeployed = probe.deployed

	if opts.Action == "apply" {
		maturity := meta.CloudMaturityFor(customer.Cloud)
		if err := previewCloudGate(maturity, customer.Cloud, opts.AcceptPreview); err != nil {
			return err
		}
		if err := ensureBackendBeforeApplyAll(opts, customer, provider, spec, opts.ConfigPath); err != nil {
			return err
		}
	}

	stack := opts.Stack
	if stack == "" {
		stack = stacks.LayerAll
	}
	targetLayers := stacks.ResolveLayers(stack)
	if opts.Action == "destroy" && stack == stacks.LayerAll {
		targetLayers = stacks.DestroyOrder()
		opts.DestroyBatch = append([]stacks.Layer(nil), targetLayers...)
	}
	if opts.Action == "destroy" && opts.DryRunDestroy {
		return destroyDryRunReport(opts, customer.NamePrefix(), targetLayers, probe.deployed)
	}

	for _, layer := range targetLayers {
		if err := applyUpstreamGuard(opts, layer, probe.deployed, opts.ConfigPath); err != nil {
			return err
		}
		if err := runOneLayer(opts, customer, meta, provider, spec, contentRoot, merged, layer); err != nil {
			return usererr.Wrap(err,
				fmt.Sprintf("OpenTofu %s failed on stack %q.", opts.Action, layer),
				"Read the OpenTofu output above, fix the cause, then re-run with the same --stack.")
		}
	}
	if opts.Action == "apply" && opts.PrintChecklist && meta.ChecklistID != "" && containsLayer(targetLayers, stacks.LayerAppPlatform) {
		if text, e := checklist.Render(meta.ChecklistID); e == nil {
			fmt.Fprintln(os.Stderr, text)
		}
	}
	if opts.Action == "apply" && containsLayer(targetLayers, stacks.LayerAppPlatform) {
		written, err := outputs.MaterializeFromArtifacts(customer.NamePrefix(), outputs.MaterializeOpts{
			AWSRegion: customer.Account.Region,
		})
		if err != nil {
			return err
		}
		for _, p := range written {
			fmt.Fprintf(os.Stderr, "Wrote %s\n", p)
		}
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

// layerSpec returns the state spec used for layer. state-backend always uses
// local state: it creates the storage the other layers keep their state in.
func layerSpec(spec cloud.StateSpec, layer stacks.Layer) cloud.StateSpec {
	if layer == stacks.LayerStateBackend {
		spec.Local = true
		spec.LocalDir = filepath.Join(paths.ArtifactsDir(spec.NamePrefix), "state")
		_ = os.MkdirAll(spec.LocalDir, 0o755)
	}
	return spec
}

// stackRelDir is stacks/<profile>/<cloud>/<layer>, relative to the infra root.
func stackRelDir(profileStackID, cloudName string, layer stacks.Layer) string {
	return filepath.Join("stacks", profileStackID, cloudName, string(layer))
}

func runOneLayer(opts Options, customer *config.CustomerConfig, meta *config.ProfileMeta, provider cloud.Cloud, spec cloud.StateSpec, contentRoot string, merged map[string]any, layer stacks.Layer) error {
	stackRel := stackRelDir(meta.StackID, customer.Cloud, layer)
	stackSrc := filepath.Join(contentRoot, "infra", stackRel)
	if st, err := os.Stat(stackSrc); err != nil || !st.IsDir() {
		return fmt.Errorf("no stack for profile %q cloud %q layer %q (expected infra/%s)", meta.StackID, customer.Cloud, layer, filepath.ToSlash(stackRel))
	}

	if opts.Action == "plan" {
		up, ready, err := upstreamReadyForPlan(customer, meta, provider, spec, contentRoot, merged, layer)
		if err != nil {
			return err
		}
		if !ready {
			fmt.Fprintf(os.Stderr,
				"Skipping plan for stack %q: upstream %q has not been applied yet (apply upstream stacks first).\n",
				layer, up)
			return nil
		}
	}
	if err := foundryApplyGuard(opts, layer, customer.NamePrefix(), merged); err != nil {
		return err
	}
	if err := destroyDependencyGuard(opts, customer.NamePrefix(), layer); err != nil {
		return err
	}
	if opts.Action == "destroy" {
		if err := provider.DestroyOrphanGuard(layer, customer.Account.Region, customer.NamePrefix(), merged); err != nil {
			return err
		}
	}

	work, err := os.MkdirTemp("", "installer-tofu-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)

	infraWork := filepath.Join(work, "infra")
	if err := copyDir(filepath.Join(contentRoot, "infra"), infraWork); err != nil {
		return err
	}
	stackWork := filepath.Join(infraWork, stackRel)

	layerMerged := maputil.DeepCopyMap(merged)
	upstream := map[string]any{}
	for _, up := range stacks.Upstream(layer) {
		ub, err := cloud.Backend(provider, spec, string(up))
		if err != nil {
			return fmt.Errorf("upstream backend %s: %w", up, err)
		}
		upstream[string(up)] = ub.RemoteState()
	}
	if len(upstream) > 0 {
		layerMerged["upstream_state"] = upstream
	}
	if err := injectAzureResourceGroupFlags(customer, provider, layerSpec(spec, layer), layer, layerMerged); err != nil {
		return err
	}
	edgeOut := filepath.Join(paths.ArtifactsDir(customer.NamePrefix()), string(stacks.LayerNetworkEdge)+"-outputs.json")
	provider.MutateDNSDelegation(context.Background(), layer, customer.Account.Region, customer.NamePrefix(), edgeOut, layerMerged)
	if err := config.WriteAutoTofuVarsJSON(stackWork, layerMerged); err != nil {
		return err
	}
	_ = config.WriteMergedConfigArtifact(customer.NamePrefix(), string(layer), layerMerged)

	backend, err := cloud.Backend(provider, layerSpec(spec, layer), string(layer))
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(stackWork, "backend.tofu"), []byte(backend.HCL()), 0o644); err != nil {
		return err
	}

	tofuPath, mirrorDir, cleanup, err := resolveTofu(contentRoot)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}

	rcPath := filepath.Join(work, "tofu.rc")
	env := append(os.Environ(), "TF_IN_AUTOMATION=1")
	if mirrorDir != "" {
		if err := writeProviderRC(rcPath, mirrorDir); err != nil {
			return err
		}
		env = append(env, "TF_CLI_CONFIG_FILE="+rcPath)
	} else {
		pc, err := pluginCacheDir(contentRoot)
		if err != nil {
			return fmt.Errorf("plugin cache: %w", err)
		}
		env = append(env, "TF_PLUGIN_CACHE_DIR="+pc)
	}

	run := func(args ...string) error {
		c := exec.Command(tofuPath, args...)
		c.Dir = stackWork
		c.Env = env
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}

	varFile := filepath.Join(stackWork, config.InstallerVarFile)
	varFileArg := "-var-file=" + varFile

	initArgs := []string{"init", "-input=false"}
	if _, err := os.Stat(filepath.Join(stackWork, ".terraform.lock.hcl")); err == nil {
		initArgs = append(initArgs, "-lockfile=readonly")
	}
	initArgs = append(initArgs, "-reconfigure")
	if err := runTofuInit(tofuPath, stackWork, env, initArgs); err != nil {
		return err
	}

	switch opts.Action {
	case "plan":
		args := []string{"plan", "-input=false", varFileArg}
		if opts.PlanOut != "" {
			args = append(args, "-out="+opts.PlanOut)
		}
		return run(args...)
	case "apply":
		planFile := filepath.Join(stackWork, "plan.bin")
		if err := run("plan", "-input=false", "-out", planFile, varFileArg); err != nil {
			return err
		}
		if !opts.AllowDestroyStateful {
			planJSON, err := tofuShowJSON(tofuPath, stackWork, env, planFile)
			if err != nil {
				return err
			}
			if err := guard.CheckPlanJSON(planJSON, guard.DefaultProtectedTypes()); err != nil {
				return err
			}
			if err := outputChangeGuard(opts, layer, planJSON, opts.layerDeployed); err != nil {
				return err
			}
		}
		if !opts.Yes {
			if err := confirmApply(string(layer)); err != nil {
				return err
			}
		}
		if err := run("apply", "-input=false", planFile); err != nil {
			return err
		}
		return captureOutputs(tofuPath, stackWork, env, customer, layer, merged, opts.ConfigPath)
	case "destroy":
		if !opts.Yes {
			if err := confirmDestroy(string(layer)); err != nil {
				return err
			}
		}
		// Confirmation (or --yes) happened above; -input=false needs -auto-approve.
		if err := run("destroy", "-input=false", "-auto-approve", varFileArg); err != nil {
			return err
		}
		return clearLayerOutputs(customer.NamePrefix(), layer)
	case "output":
		return run("output", "-json")
	default:
		return fmt.Errorf("unknown action %q", opts.Action)
	}
}

func confirmApply(layer string) error {
	fmt.Fprintf(os.Stderr, "Apply stack %q? [y/N]: ", layer)
	return readYesNo()
}

func confirmDestroy(layer string) error {
	fmt.Fprintf(os.Stderr, "Destroy stack %q? [y/N]: ", layer)
	return readYesNo()
}

func readYesNo() error {
	sc := bufio.NewScanner(os.Stdin)
	if !sc.Scan() {
		return fmt.Errorf("aborted")
	}
	ans := strings.ToLower(strings.TrimSpace(sc.Text()))
	if ans != "y" && ans != "yes" {
		return fmt.Errorf("aborted")
	}
	return nil
}

func captureOutputs(tofuPath, stackWork string, env []string, customer *config.CustomerConfig, layer stacks.Layer, merged map[string]any, configPath string) error {
	prefix := customer.NamePrefix()
	awsRegion := customer.Account.Region
	c := exec.Command(tofuPath, "output", "-json")
	c.Dir = stackWork
	c.Env = env
	out, err := c.Output()
	if err != nil {
		return nil // outputs optional
	}
	var parsed map[string]any
	if json.Unmarshal(out, &parsed) != nil {
		return nil
	}
	flat := map[string]any{}
	for k, v := range parsed {
		if vm, ok := v.(map[string]any); ok {
			flat[k] = vm["value"]
		}
	}
	b, _ := json.MarshalIndent(flat, "", "  ")
	path := filepath.Join(paths.ArtifactsDir(prefix), fmt.Sprintf("%s-outputs.json", layer))
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return err
	}
	if layer == stacks.LayerAppPlatform {
		written, err := outputs.MaterializeCWD(prefix, flat, outputs.MaterializeOpts{AWSRegion: awsRegion})
		if err != nil {
			return err
		}
		for _, p := range written {
			fmt.Fprintf(os.Stderr, "Wrote %s\n", p)
		}
	}
	if layer == stacks.LayerNetworkEdge {
		cl, err := registry.Get(customer.Cloud)
		if err != nil {
			return err
		}
		if err := cl.DelegationPauseIfNeeded(merged, configPath, prefix, flat); err != nil {
			return err
		}
	}
	return nil
}

func tofuShowJSON(tofuPath, stackWork string, env []string, planFile string) ([]byte, error) {
	c := exec.Command(tofuPath, "show", "-json", planFile)
	c.Dir = stackWork
	c.Env = env
	return c.Output()
}

