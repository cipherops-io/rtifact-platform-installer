package cli

import (
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

	notImpl := func(use string) *cobra.Command {
		return &cobra.Command{
			Use:    use,
			Args:   cobra.ExactArgs(1),
			Hidden: false,
			RunE: func(_ *cobra.Command, args []string) error {
				_ = args
				_ = kubeconfigFlag
				return deploy.NotImplemented(use)
			},
		}
	}
	grp.AddCommand(notImpl("plan [releases.yaml]"), notImpl("apply [releases.yaml]"), notImpl("status [releases.yaml]"), notImpl("destroy [releases.yaml]"))

	root.AddCommand(grp)
}