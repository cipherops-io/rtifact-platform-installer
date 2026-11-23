//go:build e2e

package deploy_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cli"
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
	if err == nil || code != 1 {
		t.Fatalf("expected refused adopt (code 1), code=%d err=%v plan:\n%s", code, err, buf.String())
	}
	if !strings.Contains(buf.String(), "--adopt") && !strings.Contains(err.Error(), "adopt") {
		t.Fatalf("expected adopt hint, plan:\n%s", buf.String())
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
	e2eAssertInventory(t, file, "ok-a", "ok-b")
}

func TestE2E_Interrupt_CancelApply(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(doc *deploy.Document) {
		doc.Defaults.Timeout = "10m"
		doc.Releases = append(doc.Releases[:1], deploy.Release{
			Name: "slow", Namespace: doc.Releases[0].Namespace, Chart: "broken", Version: "1.0.0", CRDs: "chart",
			Requires: []string{doc.Releases[0].Name},
		})
	})
	o := e2eOpts(t, file)
	e2eCleanupDeploy(t, file, o)
	ns := e2eReleaseNS(t, file, "slow")

	hc := e2eHelmClient(t)
	t.Cleanup(func() { _ = hc.Uninstall(context.Background(), "slow", ns) })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		deadline := time.Now().Add(3 * time.Minute)
		for time.Now().Before(deadline) {
			if rel, err := hc.Lookup(context.Background(), "slow", ns); err == nil && rel.Status == "pending-install" {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		cancel()
	}()
	err := deploy.Apply(ctx, o)
	e2eAssertIsInterrupted(t, err)
	e2eIntEq(t, "CLI exit code", cli.ExitCode(err), 130)
	e2eAssertLockFree(t, file)
	e2eAssertInventory(t, file, "ok-a")
}

func TestE2E_RollbackHealthyUpgrade(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, func(*deploy.Document) {})
	o := e2eOpts(t, file)
	e2eCleanupDeploy(t, file, o)
	e2eMustApply(t, o)
	nsA := e2eReleaseNS(t, file, "ok-a")
	file = writeFixture(t, dir, func(doc *deploy.Document) { doc.Releases[0].Version = "1.0.1" })
	o.FilePath = file
	e2eMustApply(t, o)
	e2eIntEq(t, "ok-a revision after upgrade", e2eHelmRevision(t, "ok-a", nsA), 2)
	if err := deploy.Rollback(context.Background(), o, "ok-a"); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	rev, status, err := e2eHelmReleaseInfo(t, "ok-a", nsA)
	if err != nil {
		t.Fatalf("ok-a must still be installed after rollback: %v", err)
	}
	if rev != 3 || status != "deployed" {
		t.Fatalf("after rollback: revision %d status %q, want 3 deployed", rev, status)
	}
	e2eAssertLockFree(t, file)
}
