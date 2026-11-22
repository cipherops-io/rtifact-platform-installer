package deploy

import (
	"context"
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
		return err
	}
	doc, disc, kcfg := sess.Doc, sess.Disc, sess.Kube
	PrintBanner(o.err(), doc, disc, kcfg, sess.Identity)
	if err := engine.ConfirmDeployRollback(o.Yes, release); err != nil {
		return err
	}
	deployment := DeploymentID(doc)
	if err := ensureSystemNS(ctx, kcfg); err != nil {
		return err
	}
	runID := newRunID()
	h, err := lock.Acquire(ctx, kcfg, deployment, lock.HolderIdentity(runID), o.ForceUnlock)
	if err != nil {
		return err
	}
	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = h.Release(releaseCtx, kcfg, deployment)
	}()
	cs, err := newHelmClient(kcfg, doc, nil, o)
	if err != nil {
		return err
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
	rel, err := cs.ReleaseByName(ctx, release, ns)
	if err != nil {
		return err
	}
	if rel.Labels == nil || rel.Labels[helmclient.LabelDeployment] != deployment {
		return fmt.Errorf("release %q in namespace %q is not managed by deploy file %q", release, ns, deployment)
	}
	rev, err := cs.PreviousDeployedRevision(ctx, release, ns)
	if err != nil {
		return err
	}
	return cs.Rollback(ctx, release, ns, rev)
}