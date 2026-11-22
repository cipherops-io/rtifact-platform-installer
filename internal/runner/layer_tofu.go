package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/engine"
	"github.com/cipherops-io/rtifact-platform-installer/internal/guard"
)

type layerTofuStep struct {
	lr *layerRun
}

func (s *layerTofuStep) Run(ctx context.Context) error {
	lr := s.lr
	if lr.planSkipped {
		return nil
	}
	initArgs := []string{"init", "-input=false"}
	if _, err := os.Stat(filepath.Join(lr.stackWork, ".terraform.lock.hcl")); err == nil {
		initArgs = append(initArgs, "-lockfile=readonly")
	}
	initArgs = append(initArgs, "-reconfigure")
	if err := runTofuInit(lr.tofuPath, lr.stackWork, lr.tofuEnv, initArgs); err != nil {
		return err
	}
	switch lr.opts.Action {
	case "plan":
		return runLayerPlan(lr)
	case "apply":
		return runLayerApply(ctx, lr)
	case "destroy":
		return runLayerDestroy(lr)
	case "output":
		return lr.tofu.Run("output", "-json")
	default:
		return fmt.Errorf("unknown action %q", lr.opts.Action)
	}
}

func runLayerPlan(lr *layerRun) error {
	args := []string{"plan", "-input=false", lr.varFileArg}
	if lr.opts.PlanOut != "" {
		args = append(args, "-out="+lr.opts.PlanOut)
	}
	return lr.tofu.Run(args...)
}

func runLayerApply(ctx context.Context, lr *layerRun) error {
	planFile := filepath.Join(lr.stackWork, "plan.bin")
	if err := lr.tofu.Run("plan", "-input=false", "-out", planFile, lr.varFileArg); err != nil {
		return err
	}
	if !lr.opts.AllowDestroyStateful {
		planJSON, err := lr.tofu.ShowJSON(planFile)
		if err != nil {
			return err
		}
		if err := guard.CheckPlanJSON(planJSON, guard.DefaultProtectedTypes()); err != nil {
			return err
		}
		if err := outputChangeGuard(lr.opts, lr.layer, planJSON, lr.opts.layerDeployed); err != nil {
			return err
		}
	}
	if err := engine.ConfirmIfNeeded(engine.Options{Yes: lr.opts.Yes}, engine.Change{Action: "apply", Stack: string(lr.layer)}); err != nil {
		return err
	}
	if err := lr.tofu.Run("apply", "-input=false", planFile); err != nil {
		return err
	}
	return captureOutputs(ctx, lr.tofu, lr.customer, lr.layer, lr.merged, lr.opts.ConfigPath)
}

func runLayerDestroy(lr *layerRun) error {
	if err := engine.ConfirmIfNeeded(engine.Options{Yes: lr.opts.Yes}, engine.Change{Action: "destroy", Stack: string(lr.layer)}); err != nil {
		return err
	}
	if err := lr.tofu.Run("destroy", "-input=false", "-auto-approve", lr.varFileArg); err != nil {
		return err
	}
	return clearLayerOutputs(lr.customer.NamePrefix(), lr.layer)
}