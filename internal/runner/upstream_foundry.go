package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
)

func requireDataPlatformForFoundry(prefix string, merged map[string]any) error {
	feats, _ := merged["features"].(map[string]any)
	if feats == nil || feats["ai_foundry"] != true {
		return nil
	}
	path := filepath.Join(config.ArtifactsDir(prefix), "data-platform-outputs.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return usererr.WithHint(
			"ai_foundry is enabled but data-platform outputs are not available.",
			"Apply data-platform first: ./dist/rtifact apply <config> --stack data-platform --yes",
			err.Error(),
		)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return fmt.Errorf("parse data-platform outputs: %w", err)
	}
	for _, k := range []string{"ai_foundry_storage_account_id", "ai_foundry_key_vault_id"} {
		if out[k] == nil || out[k] == "" {
			return usererr.WithHint(
				fmt.Sprintf("data-platform output %q is missing (data-platform may have been destroyed or not applied).", k),
				"Re-apply --stack data-platform --yes, then ai-integrations.",
			)
		}
	}
	return nil
}

func foundryApplyGuard(opts Options, layer stacks.Layer, prefix string, merged map[string]any) error {
	if opts.Action != "apply" || layer != stacks.LayerAIIntegration {
		return nil
	}
	return requireDataPlatformForFoundry(prefix, merged)
}