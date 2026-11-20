package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
	"github.com/cipherops-io/rtifact-platform-installer/internal/bootstrap"
	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud/registry"
	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/outputs"
	"github.com/cipherops-io/rtifact-platform-installer/internal/preflight"
	"github.com/cipherops-io/rtifact-platform-installer/internal/quota"
	"github.com/cipherops-io/rtifact-platform-installer/internal/runner"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
	"github.com/cipherops-io/rtifact-platform-installer/internal/status"
	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
	"github.com/cipherops-io/rtifact-platform-installer/internal/version"
	"github.com/spf13/cobra"
)

func addCloudGroups(root *cobra.Command) {
	for _, cloudName := range version.PackagedCloudsList() {
		grp := &cobra.Command{
			Use:   cloudName,
			Short: fmt.Sprintf("%s platform commands", cloudName),
		}
		addCloudLifecycle(grp, cloudName)
		root.AddCommand(grp)
	}
}

func addCloudLifecycle(grp *cobra.Command, cloudName string) {
	var stackFlag string
	var yes bool
	var allowDestroy bool
	var ignoreRuntimeDeps bool
	var dryRunDestroy bool
	var acceptPreview bool

	enforce := func(cfgCloud string) error {
		if cfgCloud != cloudName {
			return usererr.WithHint(
				"Config cloud does not match this command group.",
				fmt.Sprintf("Use `rtifact %s ...` or set cloud: %s in your YAML.", cfgCloud, cloudName),
				fmt.Sprintf("YAML has cloud: %q but you invoked the %q subcommand.", cfgCloud, cloudName),
			)
		}
		return registry.EnsureOperational(cloudName)
	}

	addFlags := func(cmd *cobra.Command) {
		cmd.Flags().StringVar(&stackFlag, "stack", "all", "state-backend|core-network|data-platform|app-platform|network-edge|ai-integrations|all")
		cmd.Flags().BoolVar(&yes, "yes", false, "Skip interactive confirmation")
		cmd.Flags().BoolVar(&allowDestroy, "allow-destroy-stateful", false, "Allow stateful resource replace/destroy")
		cmd.Flags().BoolVar(&ignoreRuntimeDeps, "ignore-runtime-deps", false, "Apply/destroy without runtime ordering checks")
		cmd.Flags().BoolVar(&acceptPreview, "accept-preview", false, "Allow apply on AWS/GCP preview clouds")
	}

	wrapRun := func(action string, checklist bool) func(*cobra.Command, []string) error {
		return func(_ *cobra.Command, args []string) error {
			cfg, err := config.LoadCustomer(args[0])
			if err != nil {
				return err
			}
			if err := enforce(cfg.Cloud); err != nil {
				return err
			}
			layer, ok := stacks.ParseLayer(stackFlag)
			if !ok {
				return usererr.WithHint(
					"Unknown stack name.",
					"Use --stack with: state-backend, core-network, data-platform, app-platform, network-edge, ai-integrations, or all.",
					fmt.Sprintf("You passed --stack %q.", stackFlag),
				)
			}
			return runner.Run(runner.Options{
				Action:               action,
				ConfigPath:           args[0],
				Stack:                layer,
				Yes:                  yes,
				AllowDestroyStateful: allowDestroy,
				PrintChecklist:       checklist,
				IgnoreRuntimeDeps:    ignoreRuntimeDeps,
				DryRunDestroy:        dryRunDestroy && action == "destroy",
				AcceptPreview:        acceptPreview,
			})
		}
	}

	for _, spec := range []struct {
		use       string
		action    string
		checklist bool
	}{
		{"plan [config.yaml]", "plan", false},
		{"apply [config.yaml]", "apply", true},
		{"destroy [config.yaml]", "destroy", false},
	} {
		cmd := &cobra.Command{Use: spec.use, Args: cobra.ExactArgs(1)}
		addFlags(cmd)
		if spec.action == "destroy" {
			cmd.Flags().BoolVar(&dryRunDestroy, "dry-run", false, "Print destroy order and dependency guard results without acting")
		}
		cmd.RunE = wrapRun(spec.action, spec.checklist)
		grp.AddCommand(cmd)
	}

	var outputTofu bool
	outputCmd := &cobra.Command{
		Use:   "output [config.yaml]",
		Short: "Show important deployment outputs (kubeconfig, bastion key, cluster IP, …)",
		Args:  cobra.ExactArgs(1),
	}
	outputCmd.Flags().BoolVar(&outputTofu, "tofu", false, "Run tofu output -json for --stack instead of the summary")
	addFlags(outputCmd)
	outputCmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadCustomer(args[0])
		if err != nil {
			return err
		}
		if err := enforce(cfg.Cloud); err != nil {
			return err
		}
		if outputTofu {
			return wrapRun("output", false)(cmd, args)
		}
		return outputs.Print(args[0])
	}
	grp.AddCommand(outputCmd)

	var checkQuota bool
	var explain bool
	validate := &cobra.Command{
		Use:   "validate [config.yaml]",
		Args:  cobra.ExactArgs(1),
		Short: "Validate config",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadCustomer(args[0])
			if err != nil {
				return err
			}
			if err := enforce(cfg.Cloud); err != nil {
				return err
			}
			checkQuota, _ = cmd.Flags().GetBool("check-quota")
			explain, _ = cmd.Flags().GetBool("explain")
			if err := config.ValidateCustomerYAML(args[0]); err != nil {
				return err
			}
			if explain {
				if err := config.PrintExplain(args[0]); err != nil {
					return err
				}
			}
			if !checkQuota {
				fmt.Fprintf(os.Stdout, "OK (skip quota; use --check-quota)\n")
				return nil
			}
			if cloudName != "azure" {
				fmt.Fprintf(os.Stdout, "OK (quota check not implemented for %s)\n", cloudName)
				return nil
			}
			root, _, err := assets.ContentRoot()
			if err != nil {
				return err
			}
			def, err := config.LoadProfileDefaults(filepath.Join(root, "profiles"), cfg.Profile)
			if err != nil {
				return err
			}
			merged, err := config.MergeBuildFromPath(def, args[0])
			if err != nil {
				return err
			}
			return quota.ValidateMerged(args[0], merged)
		},
	}
	validate.Flags().BoolVar(&checkQuota, "check-quota", cloudName == "azure", "Run cloud quota checks when available")
	validate.Flags().BoolVar(&explain, "explain", false, "Print merge provenance")
	grp.AddCommand(validate)

	grp.AddCommand(&cobra.Command{
		Use:  "preflight [config.yaml]",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg, err := config.LoadCustomer(args[0])
			if err != nil {
				return err
			}
			if err := enforce(cfg.Cloud); err != nil {
				return err
			}
			return preflight.Run(args[0])
		},
	})

	bootstrapCmd := &cobra.Command{
		Use:  "bootstrap [config.yaml]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadCustomer(args[0])
			if err != nil {
				return err
			}
			if err := enforce(cfg.Cloud); err != nil {
				return err
			}
			y, _ := cmd.Flags().GetBool("yes")
			return bootstrap.Run(args[0], y)
		},
	}
	bootstrapCmd.Flags().Bool("yes", false, "Skip confirmation")
	grp.AddCommand(bootstrapCmd)

	grp.AddCommand(&cobra.Command{
		Use:  "status [config.yaml]",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg, err := config.LoadCustomer(args[0])
			if err != nil {
				return err
			}
			if err := enforce(cfg.Cloud); err != nil {
				return err
			}
			return status.Print(args[0])
		},
	})
}
