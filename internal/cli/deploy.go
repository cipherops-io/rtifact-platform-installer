package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy"
	"github.com/spf13/cobra"
)

type exitStatus struct{ code int }

func (e *exitStatus) Error() string {
	return fmt.Sprintf("exit status %d", e.code)
}

func addDeployGroup(root *cobra.Command) {
	var kubeconfigFlag string

	grp := &cobra.Command{
		Use:   "deploy",
		Short: "Helm releases into the cluster (separate from infrastructure YAML)",
		Example: `  rtifact deploy validate examples/deploy/minimal.yaml
  rtifact deploy plan releases.yaml
  rtifact deploy plan -o json releases.yaml
  rtifact deploy apply --yes releases.yaml`,
	}
	grp.PersistentFlags().StringVar(&kubeconfigFlag, "kubeconfig", "", "Kubeconfig path (overrides file discovery)")

	validateCmd := &cobra.Command{
		Use:   "validate [releases.yaml]",
		Args:  cobra.ExactArgs(1),
		Short: "Schema and semantic validation (no cluster)",
		Example: `  rtifact deploy validate releases.yaml
  rtifact deploy validate --explain releases.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			explain, _ := cmd.Flags().GetBool("explain")
			if explain {
				return deploy.PrintExplain(args[0])
			}
			if err := deploy.ValidateFile(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "OK\n")
			return nil
		},
	}
	validateCmd.Flags().Bool("explain", false, "Print merge-style summary with ${env.*} masked")
	grp.AddCommand(validateCmd)

	writeInit := func(cmd *cobra.Command) error {
		out, _ := cmd.Flags().GetString("out")
		force, _ := cmd.Flags().GetBool("force")
		return deploy.WriteInitConfig(out, force)
	}
	initCfg := &cobra.Command{
		Use:   "init-config",
		Short: "Write a minimal deploy file",
		Example: `  rtifact deploy init-config --out releases.yaml`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return writeInit(cmd)
		},
	}
	initCfg.AddCommand(&cobra.Command{
		Use: "write",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return writeInit(cmd)
		},
	})
	initCfg.PersistentFlags().String("out", "releases.yaml", "output path")
	initCfg.PersistentFlags().Bool("force", false, "Overwrite an existing file")
	grp.AddCommand(initCfg)

	runFlags := func(cmd *cobra.Command) deploy.RunOptions {
		yes, _ := cmd.Flags().GetBool("yes")
		only, _ := cmd.Flags().GetString("only")
		det, _ := cmd.Flags().GetBool("detailed-exitcode")
		timeout, _ := cmd.Flags().GetDuration("timeout")
		return deploy.RunOptions{
			KubeconfigCLI:    kubeconfigFlag,
			Yes:              yes,
			Only:             only,
			DetailedExitCode: det,
			Timeout:          timeout,
			AllowDestroy:     mustBool(cmd, "allow-destroy"),
			AllowMassRemoval: mustBool(cmd, "allow-mass-removal"),
			AllowDowngrade:   mustBool(cmd, "allow-downgrade"),
			AllowAdopt:       mustBool(cmd, "adopt"),
			ForceUnlock:      mustBool(cmd, "force-unlock"),
			Purge:            mustBool(cmd, "purge"),
			Debug:            mustBool(cmd, "debug"),
			JSONOutput:       outputJSON(cmd),
		}
	}

	addClusterFlags := func(cmd *cobra.Command) {
		cmd.Flags().Bool("yes", false, "Skip confirmation prompts")
		cmd.Flags().String("only", "", "Only apply this release and its requires closure")
		cmd.Flags().Bool("detailed-exitcode", false, "Exit 2 when changes are pending (plan only)")
		cmd.Flags().Duration("timeout", 0, "Override default release timeout")
		cmd.Flags().Bool("allow-destroy", false, "Allow removing protected releases")
		cmd.Flags().Bool("allow-mass-removal", false, "Allow removing more than half of managed releases")
		cmd.Flags().Bool("allow-downgrade", false, "Allow chart version downgrades")
		cmd.Flags().Bool("adopt", false, "Adopt existing releases into inventory")
		cmd.Flags().Bool("force-unlock", false, "Clear a stale deploy lock")
		cmd.Flags().Bool("purge", false, "Delete PVCs and other retained resources on removal")
		cmd.Flags().Bool("debug", false, "Verbose deploy diagnostics (secrets redacted)")
	}

	addOutputFlag := func(cmd *cobra.Command) {
		cmd.Flags().StringP("output", "o", "", "Output format (json)")
	}

	planCmd := &cobra.Command{
		Use:   "plan [releases.yaml]",
		Args:  cobra.ExactArgs(1),
		Short: "Show pending cluster changes without applying",
		Example: `  rtifact deploy plan releases.yaml
  rtifact deploy plan --detailed-exitcode releases.yaml
  rtifact deploy plan -o json releases.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o := runFlags(cmd)
			o.FilePath = args[0]
			code, err := deploy.Plan(context.Background(), o)
			if err != nil {
				return err
			}
			if code != 0 {
				return &exitStatus{code: code}
			}
			return nil
		},
	}
	addClusterFlags(planCmd)
	addOutputFlag(planCmd)
	grp.AddCommand(planCmd)

	applyCmd := &cobra.Command{
		Use:   "apply [releases.yaml]",
		Args:  cobra.ExactArgs(1),
		Short: "Plan, confirm, and apply Helm releases",
		Example: `  rtifact deploy apply releases.yaml
  rtifact deploy apply --yes --timeout 15m releases.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := deploy.SignalContext(context.Background())
			defer stop()
			o := runFlags(cmd)
			o.FilePath = args[0]
			return deploy.Apply(ctx, o)
		},
	}
	addClusterFlags(applyCmd)
	grp.AddCommand(applyCmd)

	statusCmd := &cobra.Command{
		Use:   "status [releases.yaml]",
		Args:  cobra.ExactArgs(1),
		Short: "Show per-release wanted vs deployed versions",
		Example: `  rtifact deploy status releases.yaml
  rtifact deploy status -o json releases.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o := runFlags(cmd)
			o.FilePath = args[0]
			return deploy.Status(context.Background(), o)
		},
	}
	addOutputFlag(statusCmd)
	grp.AddCommand(statusCmd)

	destroyCmd := &cobra.Command{
		Use:   "destroy [releases.yaml]",
		Args:  cobra.ExactArgs(1),
		Short: "Uninstall all managed releases in reverse order",
		Example: `  rtifact deploy destroy releases.yaml
  rtifact deploy destroy --yes releases.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := deploy.SignalContext(context.Background())
			defer stop()
			o := runFlags(cmd)
			o.FilePath = args[0]
			return deploy.Destroy(ctx, o)
		},
	}
	addClusterFlags(destroyCmd)
	grp.AddCommand(destroyCmd)

	rollbackCmd := &cobra.Command{
		Use:   "rollback [releases.yaml] [release]",
		Args:  cobra.ExactArgs(2),
		Short: "Roll back one release to the previous revision",
		Example: `  rtifact deploy rollback releases.yaml my-release`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := deploy.SignalContext(context.Background())
			defer stop()
			o := runFlags(cmd)
			o.FilePath = args[0]
			return deploy.Rollback(ctx, o, args[1])
		},
	}
	rollbackCmd.Flags().Bool("yes", false, "Skip confirmation prompts")
	grp.AddCommand(rollbackCmd)

	root.AddCommand(grp)
}

func outputJSON(cmd *cobra.Command) bool {
	o, _ := cmd.Flags().GetString("output")
	return o == "json"
}

func mustBool(cmd *cobra.Command, name string) bool {
	v, _ := cmd.Flags().GetBool(name)
	return v
}

func planExitCode(err error) (int, bool) {
	var es *exitStatus
	if errors.As(err, &es) {
		return es.code, true
	}
	return 0, false
}