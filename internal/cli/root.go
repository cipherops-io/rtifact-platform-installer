package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
	"github.com/cipherops-io/rtifact-platform-installer/internal/bootstrap"
	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/dataops"
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

func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "rtifact",
		Short: "Rtifact multicloud platform installer",
	}
	root.SilenceUsage = true
	root.SilenceErrors = true

	var stackFlag string
	var yes bool
	var checkAzure bool
	var explain bool
	var allowDestroy bool
	var ignoreRuntimeDeps bool
	var dryRunDestroy bool
	var acceptPreview bool

	addStackFlags := func(cmd *cobra.Command) {
		cmd.Flags().StringVar(&stackFlag, "stack", "all", "state-backend|core-network|data-platform|app-platform|network-edge|ai-integrations|all")
		cmd.Flags().BoolVar(&yes, "yes", false, "Skip interactive confirmation")
		cmd.Flags().BoolVar(&allowDestroy, "allow-destroy-stateful", false, "Allow stateful resource replace/destroy")
		cmd.Flags().BoolVar(&ignoreRuntimeDeps, "ignore-runtime-deps", false, "Apply/destroy without runtime ordering checks")
		cmd.Flags().BoolVar(&acceptPreview, "accept-preview", false, "Allow apply on AWS/GCP preview clouds")
	}

	plan := &cobra.Command{Use: "plan [config.yaml]", Args: cobra.ExactArgs(1)}
	apply := &cobra.Command{Use: "apply [config.yaml]", Args: cobra.ExactArgs(1)}
	destroy := &cobra.Command{Use: "destroy [config.yaml]", Args: cobra.ExactArgs(1)}
	output := &cobra.Command{
		Use:   "output [config.yaml]",
		Short: "Show important deployment outputs (kubeconfig, bastion key, cluster IP, …)",
		Args:  cobra.ExactArgs(1),
	}
	var outputTofu bool
	output.Flags().BoolVar(&outputTofu, "tofu", false, "Run tofu output -json for --stack instead of the summary")
	addStackFlags(output)
	for _, c := range []*cobra.Command{plan, apply, destroy} {
		addStackFlags(c)
	}
	destroy.Flags().BoolVar(&dryRunDestroy, "dry-run", false, "Print destroy order and dependency guard results without acting")
	plan.RunE = tofuRun("plan", &stackFlag, yes, allowDestroy, false, ignoreRuntimeDeps, false, acceptPreview)
	apply.RunE = tofuRun("apply", &stackFlag, yes, allowDestroy, true, ignoreRuntimeDeps, false, acceptPreview)
	destroy.RunE = tofuRun("destroy", &stackFlag, yes, allowDestroy, false, ignoreRuntimeDeps, dryRunDestroy, acceptPreview)
	output.RunE = func(cmd *cobra.Command, args []string) error {
		if outputTofu {
			return tofuRun("output", &stackFlag, yes, allowDestroy, false, ignoreRuntimeDeps, false, acceptPreview)(cmd, args)
		}
		return outputs.Print(args[0])
	}
	root.AddCommand(plan, apply, destroy, output)
	addDeployGroup(root)
	addCloudGroups(root)

	validate := &cobra.Command{
		Use:   "validate [config.yaml]",
		Args:  cobra.ExactArgs(1),
		Short: "Validate config; optional Azure quota check",
		RunE: func(cmd *cobra.Command, args []string) error {
			checkAzure, _ = cmd.Flags().GetBool("check-azure")
			explain, _ = cmd.Flags().GetBool("explain")
			if err := config.ValidateCustomerYAML(args[0]); err != nil {
				return err
			}
			if explain {
				if err := config.PrintExplain(args[0]); err != nil {
					return err
				}
			}
			if !checkAzure {
				fmt.Fprintf(os.Stdout, "OK (skip quota; use --check-azure)\n")
				return nil
			}
			root, _, err := assets.ContentRoot()
			if err != nil {
				return err
			}
			cfg, _ := config.LoadCustomer(args[0])
			def, err := config.LoadProfileDefaults(filepath.Join(root, "profiles"), cfg.Profile, cfg.Cloud)
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
	validate.Flags().BoolVar(&checkAzure, "check-azure", true, "Check subscription vCPU quota")
	validate.Flags().BoolVar(&explain, "explain", false, "Print merge provenance")
	root.AddCommand(validate)

	root.AddCommand(&cobra.Command{
		Use:  "preflight [config.yaml]",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error { return preflight.Run(args[0]) },
	})

	bootstrapCmd := &cobra.Command{
		Use:  "bootstrap [config.yaml]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			y, _ := cmd.Flags().GetBool("yes")
			return bootstrap.Run(args[0], y)
		},
	}
	bootstrapCmd.Flags().Bool("yes", false, "Skip confirmation")
	root.AddCommand(bootstrapCmd)

	root.AddCommand(&cobra.Command{
		Use:  "status [config.yaml]",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error { return status.Print(args[0]) },
	})

	initCfg := &cobra.Command{Use: "init-config"}
	initCfg.AddCommand(&cobra.Command{
		Use: "write",
		RunE: func(cmd *cobra.Command, _ []string) error {
			tier, _ := cmd.Flags().GetString("tier")
			customer, _ := cmd.Flags().GetString("customer")
			out, _ := cmd.Flags().GetString("out")
			cloud, _ := cmd.Flags().GetString("cloud")
			return config.WriteInitConfig(cloud, tier, customer, out)
		},
	})
	initCfg.PersistentFlags().String("cloud", "azure", "cloud (azure, aws, gcp)")
	initCfg.PersistentFlags().String("tier", "sandbox", "tier name")
	initCfg.PersistentFlags().String("customer", "example", "customer code")
	initCfg.PersistentFlags().String("out", "config.yaml", "output path")
	root.AddCommand(initCfg)

	data := &cobra.Command{Use: "data"}
	data.AddCommand(&cobra.Command{
		Use:  "backup [config.yaml]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, _ := cmd.Flags().GetString("name")
			return dataops.Backup(args[0], s)
		},
	})
	data.Commands()[0].Flags().String("name", "", "backup suffix")
	for _, sub := range []struct {
		use string
		fn  func([]string) error
	}{
		{"provision [config.yaml]", func(a []string) error { return dataops.Provision(a[0], "") }},
		{"migrate [config.yaml]", func(a []string) error { return dataops.Migrate(a[0], "", "") }},
		{"retire [config.yaml]", func(a []string) error { return dataops.Retire(a[0], "", true) }},
	} {
		use := sub.use
		fn := sub.fn
		data.AddCommand(&cobra.Command{Use: use, Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error { return fn(args) }})
	}
	root.AddCommand(data)

	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("rtifact %s (commit %s) customer=%s clouds=%s\n", version.Version, version.GitCommit, version.Customer, version.PackagedClouds)
			fmt.Println(version.DeployEngineLine())
		},
	})

	return root
}

func tofuRun(action string, stackFlag *string, yes, allowDestroy, checklist, ignoreRuntimeDeps, dryRunDestroy, acceptPreview bool) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, args []string) error {
		layer, ok := stacks.ParseLayer(*stackFlag)
		if !ok {
			return usererr.WithHint(
				"Unknown stack name.",
				"Use --stack with: state-backend, core-network, data-platform, app-platform, network-edge, ai-integrations, or all.",
				fmt.Sprintf("You passed --stack %q.", *stackFlag),
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
			DryRunDestroy:        dryRunDestroy,
			AcceptPreview:        acceptPreview,
		})
	}
}

func Execute() {
	if err := NewRoot().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, usererr.Human(err))
		os.Exit(1)
	}
}
