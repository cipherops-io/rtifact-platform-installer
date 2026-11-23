package cli

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy"
	"github.com/spf13/cobra"
)

func deploySub(t *testing.T, name string) *cobra.Command {
	t.Helper()
	root := NewRoot()
	cmd, _, err := root.Find([]string{"deploy", name})
	if err != nil || cmd.Name() != name {
		t.Fatalf("deploy %s: %v", name, err)
	}
	return cmd
}

func TestDeployOutputFlag_rejectsUnknownFormat(t *testing.T) {
	for _, sub := range []string{"plan", "status"} {
		root := NewRoot()
		root.SetOut(io.Discard)
		root.SetErr(io.Discard)
		root.SetArgs([]string{"deploy", sub, "-o", "yaml", "releases.yaml"})
		err := root.Execute()
		if err == nil || !strings.Contains(err.Error(), "must be text or json") {
			t.Fatalf("%s -o yaml: want format error, got %v", sub, err)
		}
	}
}

func TestDeployOutputFlag_acceptsTextAndJSON(t *testing.T) {
	for _, v := range []string{"text", "json"} {
		cmd := deploySub(t, "status")
		if err := cmd.Flags().Set("output", v); err != nil {
			t.Fatalf("-o %s: %v", v, err)
		}
		if got := outputJSON(cmd); got != (v == "json") {
			t.Fatalf("-o %s: outputJSON=%v", v, got)
		}
	}
}

func TestDeployRollback_hasLockAndTimeoutFlags(t *testing.T) {
	cmd := deploySub(t, "rollback")
	for _, f := range []string{"yes", "timeout", "force-unlock"} {
		if cmd.Flags().Lookup(f) == nil {
			t.Errorf("rollback missing --%s", f)
		}
	}
}

func TestExitCode_wrappedInterruptIs130(t *testing.T) {
	if got := ExitCode(deploy.AsInterrupted(context.Canceled)); got != 130 {
		t.Fatalf("exit code %d, want 130", got)
	}
}
