package deploy

import (
	"context"
	"fmt"
	"time"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/lock"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
	"github.com/cipherops-io/rtifact-platform-installer/internal/engine"
)

// Destroy removes managed releases in reverse dependency order (S18).
func Destroy(ctx context.Context, o RunOptions) error {
	sess, err := OpenSession(ctx, o)
	if err != nil {
		return AsInterrupted(err)
	}
	doc, disc, kcfg := sess.Doc, sess.Disc, sess.Kube
	PrintBanner(o.err(), doc, disc, kcfg, sess.Identity)
	deployment := DeploymentID(doc)
	if err := engine.ConfirmDeployDestroy(o.Yes, deployment); err != nil {
		return AsInterrupted(err)
	}
	if err := ensureSystemNS(ctx, kcfg); err != nil {
		return AsInterrupted(err)
	}
	runID := newRunID()
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	h, err := lock.Acquire(runCtx, kcfg, deployment, lock.HolderIdentity(runID), o.ForceUnlock)
	if err != nil {
		return AsInterrupted(err)
	}
	h.OnRenewFailure(func(err error) {
		fmt.Fprintf(o.err(), "deploy lock renewal failed: %v\n", err)
		cancel()
	})
	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = h.Release(releaseCtx, kcfg, deployment)
	}()
	inv, err := inventory.Read(runCtx, kcfg, deployment)
	if err != nil {
		return AsInterrupted(err)
	}
	for _, e := range inv {
		if e.Protect && !o.AllowDestroy {
			return fmt.Errorf("%s/%s: protected release (use --allow-destroy)", e.Namespace, e.Name)
		}
	}
	if len(inv) == 0 {
		fmt.Fprintln(o.out(), "No managed releases in inventory.")
		return nil
	}
	cs, err := newHelmClient(kcfg, doc, nil, o)
	if err != nil {
		return AsInterrupted(err)
	}
	order := plan.RemovalOrder(inv)
	remaining := map[string]inventory.Entry{}
	for _, e := range inv {
		remaining[entryKey(e.Name, e.Namespace)] = e
	}
	for _, e := range order {
		fmt.Fprintf(o.out(), "  uninstall %s/%s\n", e.Namespace, e.Name)
		manifest, err := fetchReleaseManifest(runCtx, cs, e.Name, e.Namespace)
		if err != nil {
			return AsInterrupted(err)
		}
		if err := cs.Uninstall(runCtx, e.Name, e.Namespace); err != nil {
			return AsInterrupted(err)
		}
		delete(remaining, entryKey(e.Name, e.Namespace))
		if err := inventory.Write(runCtx, kcfg, deployment, inventoryFromMap(remaining)); err != nil {
			return AsInterrupted(err)
		}
		opt := purgeRemovalOpts{ReleaseName: e.Name, ReleaseNS: e.Namespace, Manifest: manifest}
		if err := reportOrPurgeRemovalData(runCtx, kcfg, doc, remaining, opt, o.Purge, o.out()); err != nil {
			return AsInterrupted(err)
		}
	}
	return nil
}