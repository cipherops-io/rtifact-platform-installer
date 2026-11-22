//go:build e2e

package e2e_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// azureApplyOrder matches stacks.ApplyOrder plus optional state-backend (bootstrap).
var azureApplyOrder = []string{
	"state-backend",
	"core-network",
	"data-platform",
	"app-platform",
	"network-edge",
	"ai-integrations",
}

func rtifactBin(t *testing.T) (bin, contentRoot string) {
	t.Helper()
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	contentRoot = filepath.Join(root, "..", "..")
	bin = filepath.Join(contentRoot, "dist", "rtifact")
	if _, err := os.Stat(bin); err != nil {
		t.Fatal("build rtifact first: make build-dev PACKAGED_CLOUDS=azure")
	}
	return bin, contentRoot
}

func runRtifact(t *testing.T, bin, contentRoot string, args ...string) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(),
		"INSTALLER_CONTENT_ROOT="+contentRoot,
		"ARM_USE_OIDC=true",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("rtifact %v:\n%s", args, out)
	}
}

// TestAzureSmoke runs validate + preflight (no Azure mutations).
func TestAzureSmoke(t *testing.T) {
	cfg := os.Getenv("E2E_AZURE_CONFIG")
	if cfg == "" {
		t.Skip("E2E_AZURE_CONFIG not set")
	}
	bin, root := rtifactBin(t)
	check := os.Getenv("E2E_CHECK_AZURE_QUOTA")
	if check == "" {
		check = "false"
	}
	runRtifact(t, bin, root, "validate", cfg, "--check-quota="+check)
	runRtifact(t, bin, root, "preflight", cfg)
}

// TestAzureStackAction runs plan, apply, or destroy for one or all Azure stacks.
// Env:
//   - E2E_AZURE_CONFIG (required)
//   - E2E_AZURE_ACTION: plan | apply | destroy (default plan)
//   - E2E_AZURE_STACK: layer name or all (default core-network)
//   - E2E_AZURE_BOOTSTRAP: if true, run bootstrap before apply (remote state)
func TestAzureStackAction(t *testing.T) {
	cfg := os.Getenv("E2E_AZURE_CONFIG")
	if cfg == "" {
		t.Skip("E2E_AZURE_CONFIG not set")
	}
	action := strings.ToLower(strings.TrimSpace(os.Getenv("E2E_AZURE_ACTION")))
	if action == "" {
		action = "plan"
	}
	stack := strings.TrimSpace(os.Getenv("E2E_AZURE_STACK"))
	if stack == "" {
		stack = "core-network"
	}
	bin, root := rtifactBin(t)

	if os.Getenv("E2E_AZURE_BOOTSTRAP") == "true" && action == "apply" {
		runRtifact(t, bin, root, "bootstrap", cfg, "--yes")
	}

	layers := layersForStack(stack, action)
	for _, layer := range layers {
		switch action {
		case "plan":
			runRtifact(t, bin, root, "plan", cfg, "--stack", layer, "--yes")
		case "apply":
			runRtifact(t, bin, root, "apply", cfg, "--stack", layer, "--yes")
		case "destroy":
			runRtifact(t, bin, root, "destroy", cfg, "--stack", layer, "--yes")
		default:
			t.Fatalf("unknown E2E_AZURE_ACTION %q", action)
		}
	}
}

func layersForStack(stack, action string) []string {
	if stack != "all" {
		return []string{stack}
	}
	order := make([]string, len(azureApplyOrder))
	copy(order, azureApplyOrder)
	if action == "destroy" {
		for i, j := 0, len(order)-1; i < j; i, j = i+1, j-1 {
			order[i], order[j] = order[j], order[i]
		}
	}
	return order
}