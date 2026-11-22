//go:build e2e

package deploy_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy"
)

func TestE2E_Adopt_UnmanagedRelease(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(*deploy.Document) {})
	o := e2eOpts(t, file)
	e2eCleanupDeploy(t, file, o)
	nsA := e2eReleaseNS(t, file, "ok-a")
	e2eHelmUpgradeNoWait(t, "ok-a", nsA, "ok-a", "1.0.0")
	var buf bytes.Buffer
	oo := e2eOpts(t, file)
	oo.Stdout = &buf
	code, err := deploy.Plan(context.Background(), oo)
	if err != nil {
		t.Fatal(err)
	}
	if code != 1 || !strings.Contains(buf.String(), "--adopt") {
		t.Fatalf("expected refused adopt hint, code=%d plan:\n%s", code, buf.String())
	}
	oo.AllowAdopt = true
	oo.DetailedExitCode = true
	buf.Reset()
	code, err = deploy.Plan(context.Background(), oo)
	if err != nil {
		t.Fatal(err)
	}
	e2eIntEq(t, "plan adopt exit", code, 2)
	if !strings.Contains(buf.String(), "adopt") {
		t.Fatalf("expected adopt in plan:\n%s", buf.String())
	}
	o.AllowAdopt = true
	e2eMustApply(t, o)
	e2eAssertInventory(t, file, "ok-a")
}

func TestE2E_Interrupt_CancelApply(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Defaults.Timeout = "10m"
	})
	o := e2eOpts(t, file)
	e2eCleanupDeploy(t, file, o)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		time.Sleep(300 * time.Millisecond)
		cancel()
	}()
	err := deploy.Apply(ctx, o)
	e2eAssertIsInterrupted(t, err)
}