package deploy

import (
	"context"
	"errors"
	"fmt"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient"
	"github.com/cipherops-io/rtifact-platform-installer/internal/engine"
)

// Rollback rolls back a single release to its previous deployed revision. A
// release with no deployed revision (pending or failed first install) has its
// Helm record removed instead; PVCs are kept by the keep post-renderer.
func Rollback(ctx context.Context, o RunOptions, release string) error {
	ctx, stop := SignalContext(ctx)
	defer stop()
	sess, err := OpenSession(ctx, o)
	if err != nil {
		return AsInterrupted(err)
	}
	PrintBanner(o.err(), sess.Doc, sess.Disc, sess.Kube, sess.Identity)
	ns := releaseNamespace(sess.Doc, release)
	if ns == "" {
		return fmt.Errorf("release %q not in deploy file", release)
	}
	if err := engine.ConfirmDeployRollback(o.Yes, release); err != nil {
		return AsInterrupted(err)
	}
	err = withLock(ctx, sess, o, "rollback", func(ctx context.Context) error {
		hc, err := sess.helm(o, nil)
		if err != nil {
			return err
		}
		return rollbackRelease(ctx, o, hc, sess.DeploymentID(), release, ns)
	})
	return AsInterrupted(err)
}

func releaseNamespace(doc *Document, release string) string {
	for _, r := range doc.Releases {
		if r.Name == release {
			return r.Namespace
		}
	}
	return ""
}

func rollbackRelease(ctx context.Context, o RunOptions, hc helmclient.Interface, deployment, release, ns string) error {
	rel, err := hc.Lookup(ctx, release, ns)
	if err != nil {
		return err
	}
	if rel.Labels[helmclient.LabelDeployment] != deployment {
		return fmt.Errorf("release %q in namespace %q is not managed by deploy file %q", release, ns, deployment)
	}
	if rel.PendingInstall() {
		return clearReleaseRecord(ctx, o, hc, release, ns, rel.Status)
	}
	rev, err := hc.PreviousDeployedRevision(ctx, release, ns)
	if errors.Is(err, helmclient.ErrNoPreviousDeployedRevision) {
		return clearReleaseRecord(ctx, o, hc, release, ns, rel.Status)
	}
	if err != nil {
		return err
	}
	return hc.Rollback(ctx, release, ns, rev)
}

func clearReleaseRecord(ctx context.Context, o RunOptions, hc helmclient.Interface, release, ns, status string) error {
	fmt.Fprintf(o.out(), "Release %q in namespace %q is %s with no deployed revision to roll back to.\n", release, ns, status)
	fmt.Fprintln(o.out(), "Uninstalling the release record (PVCs are kept).")
	if err := hc.Uninstall(ctx, release, ns); err != nil {
		return err
	}
	fmt.Fprintf(o.out(), "Recovery: run rtifact deploy apply %s to reconcile.\n", o.FilePath)
	return nil
}
