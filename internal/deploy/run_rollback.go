package deploy

import (
	"context"
	"fmt"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/ocireg"
	"github.com/cipherops-io/rtifact-platform-installer/internal/engine"
)

// Rollback rolls back a single release to its previous revision.
func Rollback(ctx context.Context, o RunOptions, release string) error {
	doc, disc, kcfg, err := loadRuntime(o)
	if err != nil {
		return err
	}
	printBanner(o.out(), doc, disc, kcfg)
	if err := engine.ConfirmDeployRollback(o.Yes, release); err != nil {
		return err
	}
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
		Kube: kcfg, Deployment: DeploymentID(doc), RegistryClient: reg,
		Defaults: helmclient.Defaults{
			Timeout: doc.Defaults.Timeout, Wait: doc.Defaults.Wait,
			RollbackOnFailure: doc.Defaults.RollbackOnFailure, HistoryMax: doc.Defaults.HistoryMax,
		},
		Registry: helmclient.RegistryOpts{PlainHTTP: doc.Registry.PlainHTTP},
	})
	if err != nil {
		return err
	}
	rel, err := cs.ReleaseByName(release)
	if err != nil {
		return err
	}
	rev := rel.Version - 1
	if rev < 1 {
		return fmt.Errorf("release %q has no previous revision", release)
	}
	return cs.Rollback(ctx, release, rev)
}