package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy"
	"github.com/spf13/cobra"
)

func addDeployGroup(root *cobra.Command) {
	var kubeconfigFlag string

	grp := &cobra.Command{
		Use:   "deploy",
		Short: "Helm releases into the cluster (separate from infrastructure YAML)",
	}
	grp.PersistentFlags().StringVar(&kubeconfigFlag, "kubeconfig", "", "Kubeconfig path (overrides file discovery)")

	validateCmd := &cobra.Command{
		Use:   "validate [releases.yaml]",
		Args:  cobra.ExactArgs(1),
		Short: "Schema and semantic validation (no cluster)",
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

	initCfg := &cobra.Command{
		Use:   "init-config",
		Short: "Write a minimal deploy file",
	}
	initCfg.AddCommand(&cobra.Command{
		Use: "write",
		RunE: func(cmd *cobra.Command, _ []string) error {
			out, _ := cmd.Flags().GetString("out")
			return deploy.WriteInitConfig(out)
		},
	})
	initCfg.PersistentFlags().String("out", "releases.yaml", "output path")
	grp.AddCommand(initCfg)

	runFlags := func(cmd *cobra.Command) deploy.RunOptions {
		yes, _ := cmd.Flags().GetBool("yes")
		only, _ := cmd.Flags().GetString("only")
		det, _ := cmd.Flags().GetBool("detailed-exitcode")
		timeout, _ := cmd.Flags().GetDuration("timeout")
		return deploy.RunOptions{
			KubeconfigCLI: kubeconfigFlag,
			Yes:           yes,
			Only:          only,
			DetailedExitCode: det,
			Timeout:       timeout,
			AllowDestroy:      mustBool(cmd, "allow-destroy"),
			AllowMassRemoval:  mustBool(cmd, "allow-mass-removal"),
			AllowDowngrade:    mustBool(cmd, "allow-downgrade"),
			AllowAdopt:        mustBool(cmd, "allow-adopt"),
			ForceUnlock:       mustBool(cmd, "force-unlock"),
			Purge:             mustBool(cmd, "purge"),
			Debug:             mustBool(cmd, "debug"),
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
		cmd.Flags().Bool("allow-adopt", false, "Adopt existing releases into inventory")
		cmd.Flags().Bool("force-unlock", false, "Clear a stale deploy lock")
		cmd.Flags().Bool("purge", false, "Delete PVCs and other retained resources on removal")
		cmd.Flags().Bool("debug", false, "Verbose deploy diagnostics (secrets redacted)")
	}

	planCmd := &cobra.Command{
		Use:   "plan [releases.yaml]",
		Args:  cobra.ExactArgs(1),
		Short: "Show pending cluster changes without applying",
		RunE: func(cmd *cobra.Command, args []string) error {
			o := runFlags(cmd)
			o.FilePath = args[0]
			code, err := deploy.Plan(context.Background(), o)
			if err != nil {
				return err
			}
			if code == 2 {
				os.Exit(2)
			}
			return nil
		},
	}
	addClusterFlags(planCmd)
	grp.AddCommand(planCmd)

	applyCmd := &cobra.Command{
		Use:   "apply [releases.yaml]",
		Args:  cobra.ExactArgs(1),
		Short: "Plan, confirm, and apply Helm releases",
		RunE: func(cmd *cobra.Command, args []string) error {
			o := runFlags(cmd)
			o.FilePath = args[0]
			return deploy.Apply(context.Background(), o)
		},
	}
	addClusterFlags(applyCmd)
	grp.AddCommand(applyCmd)

	grp.AddCommand(&cobra.Command{
		Use:    "status [releases.yaml]",
		Args:   cobra.ExactArgs(1),
		Hidden: false,
		RunE: func(_ *cobra.Command, args []string) error {
			return deploy.NotImplemented("status " + args[0])
		},
	})
	grp.AddCommand(&cobra.Command{
		Use:    "destroy [releases.yaml]",
		Args:   cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return deploy.NotImplemented("destroy " + args[0])
		},
	})

	root.AddCommand(grp)
}

func mustBool(cmd *cobra.Command, name string) bool {
	v, _ := cmd.Flags().GetBool(name)
	return v
}