package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud/registry"
	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/outputs"
	"github.com/cipherops-io/rtifact-platform-installer/internal/paths"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

func captureOutputs(ctx context.Context, tofu TofuRunner, customer *config.CustomerConfig, layer stacks.Layer, merged map[string]any, configPath string) error {
	_ = ctx
	prefix := customer.NamePrefix()
	awsRegion := customer.Account.Region
	out, err := tofu.Output("output", "-json")
	if err != nil {
		return nil
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
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return err
	}
	if slices.Contains(outputs.MaterializeLayers, layer) {
		written, err := outputs.MaterializeCWD(prefix, flat, outputs.MaterializeOpts{AWSRegion: awsRegion})
		if err != nil {
			return err
		}
		for _, p := range written {
			fmt.Fprintf(os.Stderr, "Wrote %s\n", p)
		}
		if cwd, err := os.Getwd(); err == nil {
			if cmd := outputs.BastionSSHCommand(prefix, cwd, flat); cmd != "" {
				fmt.Fprintf(os.Stderr, "Bastion: %s\n", cmd)
			}
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