package deploy

import (
	"context"
	"fmt"
	"io"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/cluster"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
	"github.com/cipherops-io/rtifact-platform-installer/internal/engine"
)

// Destroy removes managed releases in reverse dependency order (S18).
func Destroy(ctx context.Context, o RunOptions) error {
	ctx, stop := SignalContext(ctx)
	defer stop()
	sess, err := OpenSession(ctx, o)
	if err != nil {
		return AsInterrupted(err)
	}
	PrintBanner(o.err(), sess.Doc, sess.Disc, sess.Kube, sess.Identity)
	if err := engine.ConfirmDeployDestroy(o.Yes, sess.DeploymentID()); err != nil {
		return AsInterrupted(err)
	}
	err = withLock(ctx, sess, o, "destroy", func(ctx context.Context) error {
		hc, err := sess.helm(o, nil)
		if err != nil {
			return err
		}
		return destroyRun{
			cs: sess.Client, helm: hc, doc: sess.Doc, deployment: sess.DeploymentID(),
			allowDestroy: o.AllowDestroy, purge: o.Purge, out: o.out(),
		}.run(ctx)
	})
	return AsInterrupted(err)
}

// destroyRun uninstalls every inventory entry, rewriting inventory after each
// uninstall so an interrupted destroy resumes with only what is left.
type destroyRun struct {
	cs           cluster.Client
	helm         helmclient.Interface
	doc          *Document
	deployment   string
	allowDestroy bool
	purge        bool
	out          io.Writer
}

func (d destroyRun) run(ctx context.Context) error {
	inv, err := inventory.Read(ctx, d.cs, d.deployment)
	if err != nil {
		return err
	}
	for _, e := range inv {
		if e.Protect && !d.allowDestroy {
			return fmt.Errorf("%s/%s: protected release (use --allow-destroy)", e.Namespace, e.Name)
		}
	}
	if len(inv) == 0 {
		fmt.Fprintln(d.out, "No managed releases in inventory.")
		return nil
	}
	remaining := map[string]inventory.Entry{}
	for _, e := range inv {
		remaining[entryKey(e.Name, e.Namespace)] = e
	}
	for _, e := range plan.RemovalOrder(inv) {
		if err := d.uninstall(ctx, e, remaining); err != nil {
			return err
		}
	}
	return nil
}

func (d destroyRun) uninstall(ctx context.Context, e inventory.Entry, remaining map[string]inventory.Entry) error {
	fmt.Fprintf(d.out, "  uninstall %s/%s\n", e.Namespace, e.Name)
	manifest, err := d.helm.ReleaseManifest(ctx, e.Name, e.Namespace)
	if err != nil {
		return err
	}
	if err := d.helm.Uninstall(ctx, e.Name, e.Namespace); err != nil {
		return err
	}
	delete(remaining, entryKey(e.Name, e.Namespace))
	if err := inventory.Write(ctx, d.cs, d.deployment, inventoryFromMap(remaining)); err != nil {
		return err
	}
	opt := purgeRemovalOpts{ReleaseName: e.Name, ReleaseNS: e.Namespace, Manifest: manifest}
	return reportOrPurgeRemovalData(ctx, d.cs, d.doc, remaining, opt, d.purge, d.out)
}
