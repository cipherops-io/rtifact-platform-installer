package deploy

import (
	"context"
	"fmt"
	"time"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/lock"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/ocireg"
	"github.com/cipherops-io/rtifact-platform-installer/internal/engine"
)

// Destroy removes managed releases in reverse dependency order (S18).
func Destroy(ctx context.Context, o RunOptions) error {
	sess, err := OpenSession(ctx, o)
	if err != nil {
		return err
	}
	doc, disc, kcfg := sess.Doc, sess.Disc, sess.Kube
	printBanner(o.out(), doc, disc, kcfg)
	deployment := DeploymentID(doc)
	if err := engine.ConfirmDeployDestroy(o.Yes, deployment); err != nil {
		return err
	}
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
	regCfg := registryFromDoc(doc)
	cred, err := ocireg.LoadAuth(regCfg)
	if err != nil {
		return err
	}
	reg, err := ocireg.NewClient(regCfg, cred)
	if err != nil {
		return err
	}
	cs, err := helmclient.New(helmclient.Options{
		Kube: kcfg, Deployment: deployment, RegistryClient: reg,
		Defaults: helmclient.Defaults{
			Timeout: doc.Defaults.Timeout, Wait: doc.Defaults.Wait,
			RollbackOnFailure: doc.Defaults.RollbackOnFailure, HistoryMax: doc.Defaults.HistoryMax,
		},
		Registry: helmclient.RegistryOpts{PlainHTTP: doc.Registry.PlainHTTP},
	})
	if err != nil {
		return err
	}
	order, err := RemoveOrder(doc.Releases)
	if err != nil {
		return err
	}
	relByName := map[string]Release{}
	for _, r := range doc.Releases {
		relByName[r.Name] = r
	}
	for _, name := range order {
		rel := relByName[name]
		fmt.Fprintf(o.out(), "  uninstall %s/%s\n", rel.Namespace, rel.Name)
		if err := cs.Uninstall(ctx, rel.Name, rel.Namespace); err != nil {
			return err
		}
	}
	_ = inventory.Write(ctx, kcfg, deployment, nil)
	return nil
}