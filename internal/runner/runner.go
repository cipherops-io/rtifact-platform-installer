package runner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
	"github.com/cipherops-io/rtifact-platform-installer/internal/checklist"
	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/guard"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

type Options struct {
	Action               string
	ConfigPath           string
	Stack                stacks.Layer
	Yes                  bool
	AllowDestroyStateful bool
	PrintChecklist       bool
	PlanOut              string
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
	defaults, err := config.LoadProfileDefaults(profilesDir, customer.Profile)
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

	targetLayers := stacks.ResolveLayers(opts.Stack)
	if opts.Stack == "" {
		targetLayers = []stacks.Layer{stacks.LayerNetwork}
	}
	if opts.Action == "destroy" {
		targetLayers = stacks.DestroyOrder()
		if opts.Stack != "" && opts.Stack != stacks.LayerAll {
			targetLayers = []stacks.Layer{opts.Stack}
		}
	}

	for _, layer := range targetLayers {
		if err := runOneLayer(opts, customer, meta, contentRoot, merged, layer); err != nil {
			return fmt.Errorf("stack %s: %w", layer, err)
		}
		if opts.Action == "apply" && opts.PrintChecklist && layer == stacks.LayerCluster {
			if meta.ChecklistID != "" {
				if text, e := checklist.Render(meta.ChecklistID); e == nil {
					fmt.Fprintln(os.Stderr, text)
				}
			}
		}
	}
	return nil
}

func runOneLayer(opts Options, customer *config.CustomerConfig, meta *config.ProfileMeta, contentRoot string, merged map[string]any, layer stacks.Layer) error {
	stackRel := filepath.Join("stacks", meta.StackID, string(layer))
	stackSrc := filepath.Join(contentRoot, "infra", "stacks", meta.StackID, string(layer))
	if st, err := os.Stat(stackSrc); err != nil || !st.IsDir() {
		stackSrc = filepath.Join(contentRoot, "infra", "stacks", meta.StackID)
		stackRel = filepath.Join("stacks", meta.StackID)
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

	layerMerged := deepCopyMap(merged)
	if inputs, err := loadLayerInputs(customer.NamePrefix(), layer); err == nil && inputs != nil {
		layerMerged["layer_inputs"] = inputs
	}
	if err := config.WriteAutoTofuVarsJSON(stackWork, layerMerged); err != nil {
		return err
	}
	_ = config.WriteMergedConfigArtifact(customer.NamePrefix(), string(layer), layerMerged)

	backend := config.EffectiveBackend(customer, meta.StackID, string(layer))
	if backend != nil {
		if backend.ResourceGroupName == "" || backend.StorageAccountName == "" {
			return fmt.Errorf("remote backend requires storage account configuration; run installer bootstrap first")
		}
		if err := writeBackend(stackWork, backend, customer.Azure); err != nil {
			return err
		}
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
		pc := filepath.Join(work, ".plugin-cache")
		_ = os.MkdirAll(pc, 0o755)
		env = append(env, "TF_PLUGIN_CACHE_DIR="+pc)
	}

	cwd, _ := os.Getwd()
	var stateArgs []string
	var extraInit []string
	if backend == nil {
		stateFile := filepath.Join(cwd, fmt.Sprintf("tofu-%s.tfstate", layer))
		stateArgs = []string{"-state=" + stateFile}
	} else {
		extraInit = append(extraInit, "-reconfigure")
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
	initArgs = append(initArgs, extraInit...)
	if err := run(initArgs...); err != nil {
		return err
	}

	switch opts.Action {
	case "plan":
		args := []string{"plan", "-input=false", varFileArg}
		if opts.PlanOut != "" {
			args = append(args, "-out="+opts.PlanOut)
		}
		args = append(args, stateArgs...)
		return run(args...)
	case "apply":
		planFile := filepath.Join(stackWork, "plan.bin")
		if err := run(append([]string{"plan", "-input=false", "-out", planFile, varFileArg}, stateArgs...)...); err != nil {
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
		}
		if !opts.Yes {
			if err := confirmApply(string(layer)); err != nil {
				return err
			}
		}
		if err := run(append([]string{"apply", "-input=false", planFile}, stateArgs...)...); err != nil {
			return err
		}
		return captureOutputs(tofuPath, stackWork, env, customer.NamePrefix(), layer, stateArgs)
	case "destroy":
		if !opts.Yes {
			if err := confirmDestroy(string(layer)); err != nil {
				return err
			}
		}
		args := []string{"destroy", "-input=false", varFileArg}
		if opts.Yes {
			args = append(args, "-auto-approve")
		}
		args = append(args, stateArgs...)
		return run(args...)
	case "output":
		return run(append([]string{"output", "-json"}, stateArgs...)...)
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

func layerArtifactsDir(prefix string) string {
	dir := filepath.Join(os.Getenv("HOME"), ".rtifact-installer", prefix)
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

func loadLayerInputs(prefix string, layer stacks.Layer) (map[string]any, error) {
	path := filepath.Join(layerArtifactsDir(prefix), fmt.Sprintf("%s-outputs.json", layer))
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func captureOutputs(tofuPath, stackWork string, env []string, prefix string, layer stacks.Layer, stateArgs []string) error {
	c := exec.Command(tofuPath, append([]string{"output", "-json"}, stateArgs...)...)
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
	path := filepath.Join(layerArtifactsDir(prefix), fmt.Sprintf("%s-outputs.json", layer))
	return os.WriteFile(path, b, 0o644)
}

func tofuShowJSON(tofuPath, stackWork string, env []string, planFile string) ([]byte, error) {
	c := exec.Command(tofuPath, "show", "-json", planFile)
	c.Dir = stackWork
	c.Env = env
	return c.Output()
}

func deepCopyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		if m, ok := v.(map[string]any); ok {
			out[k] = deepCopyMap(m)
		} else {
			out[k] = v
		}
	}
	return out
}