package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type layerGuardStep struct {
	lr *layerRun
}

func (s *layerGuardStep) Run(ctx context.Context) error {
	lr := s.lr
	stackRel := stackRelDir(lr.meta.StackID, lr.customer.Cloud, lr.layer)
	stackSrc := filepath.Join(lr.contentRoot, "infra", stackRel)
	if st, err := os.Stat(stackSrc); err != nil || !st.IsDir() {
		return fmt.Errorf("no stack for profile %q cloud %q layer %q (expected infra/%s)", lr.meta.StackID, lr.customer.Cloud, lr.layer, filepath.ToSlash(stackRel))
	}
	if lr.opts.Action == "plan" {
		up, ready, err := upstreamReadyForPlan(ctx, lr.customer, lr.meta, lr.provider, lr.spec, lr.contentRoot, lr.merged, lr.layer)
		if err != nil {
			return err
		}
		if !ready {
			fmt.Fprintf(os.Stderr,
				"Skipping plan for stack %q: upstream %q has not been applied yet (apply upstream stacks first).\n",
				lr.layer, up)
			lr.planSkipped = true
			return nil
		}
	}
	if err := foundryApplyGuard(lr.opts, lr.layer, lr.customer.NamePrefix(), lr.merged); err != nil {
		return err
	}
	if err := destroyDependencyGuard(lr.opts, lr.customer.NamePrefix(), lr.layer); err != nil {
		return err
	}
	if lr.opts.Action == "destroy" {
		if err := lr.provider.DestroyOrphanGuard(lr.layer, lr.customer.Account.Region, lr.customer.NamePrefix(), lr.merged); err != nil {
			return err
		}
	}
	return nil
}