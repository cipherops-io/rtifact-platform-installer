package deploy

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/lock"
	"github.com/cipherops-io/rtifact-platform-installer/internal/engine"
)

// Rollback rolls back a single release to its previous deployed revision.
func Rollback(ctx context.Context, o RunOptions, release string) error {
	sess, err := OpenSession(ctx, o)
	if err != nil {
		return AsInterrupted(err)
	}
	doc, disc, kcfg := sess.Doc, sess.Disc, sess.Kube
	PrintBanner(o.err(), doc, disc, kcfg, sess.Identity)
	if err := engine.ConfirmDeployRollback(o.Yes, release); err != nil {
		return AsInterrupted(err)
	}
	deployment := DeploymentID(doc)
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
	cs, err := newHelmClient(kcfg, doc, nil, o)
	if err != nil {
		return AsInterrupted(err)
	}
	var ns string
	for _, r := range doc.Releases {
		if r.Name == release {
			ns = r.Namespace
			break
		}
	}
	if ns == "" {
		return fmt.Errorf("release %q not in deploy file", release)
	}
	rel, err := cs.ReleaseByName(runCtx, release, ns)
	if err != nil {
		return AsInterrupted(err)
	}
	if rel.Labels == nil || rel.Labels[helmclient.LabelDeployment] != deployment {
		return fmt.Errorf("release %q in namespace %q is not managed by deploy file %q", release, ns, deployment)
	}
	if helmclient.IsPendingInstall(rel) {
		return AsInterrupted(clearReleaseRecord(runCtx, o, cs, release, ns, string(rel.Info.Status)))
	}
	rev, err := cs.PreviousDeployedRevision(runCtx, release, ns)
	if err != nil {
		if errors.Is(err, helmclient.ErrNoPreviousDeployedRevision) {
			return AsInterrupted(clearReleaseRecord(runCtx, o, cs, release, ns, string(rel.Info.Status)))
		}
		return AsInterrupted(err)
	}
	return AsInterrupted(cs.Rollback(runCtx, release, ns, rev))
}

func clearReleaseRecord(ctx context.Context, o RunOptions, cs *helmclient.Client, release, ns, status string) error {
	fmt.Fprintf(o.out(), "Release %q in namespace %q is %s with no deployed revision to roll back to.\n", release, ns, status)
	fmt.Fprintln(o.out(), "Uninstalling the release record (PVCs are kept).")
	if err := cs.Uninstall(ctx, release, ns); err != nil {
		return err
	}
	fmt.Fprintf(o.out(), "Recovery: run rtifact deploy apply %s to reconcile.\n", o.FilePath)
	return nil
}