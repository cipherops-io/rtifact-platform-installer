package deploy

import (
	"context"
	"fmt"
	"time"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/crds"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/ocireg"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
	"github.com/cipherops-io/rtifact-platform-installer/internal/engine"
)

type applyRun struct {
	o    RunOptions
	sess *Session
	*planRun

	wantByKey map[string]plan.Wanted
	entries   map[string]inventory.Entry
}

func entryKey(name, ns string) string { return ns + "/" + name }

// Apply installs or upgrades releases. The deploy lock is held before the
// cluster is observed so the plan cannot go stale under a concurrent run.
func Apply(ctx context.Context, o RunOptions) error {
	ctx, stop := SignalContext(ctx)
	defer stop()
	defer SetActiveRedactor(nil)

	sess, err := OpenSession(ctx, o)
	if err != nil {
		return AsInterrupted(err)
	}
	PrintBanner(o.err(), sess.Doc, sess.Disc, sess.Kube, sess.Identity)
	err = withLock(ctx, sess, o, "apply", func(ctx context.Context) error {
		pr, err := buildPlan(ctx, sess, &o)
		if err != nil {
			return err
		}
		if err := pr.print(ctx, sess, o); err != nil {
			return err
		}
		return newApplyRun(o, sess, pr).run(ctx)
	})
	return AsInterrupted(RedactError(ActiveRedactor(), err))
}

func newApplyRun(o RunOptions, sess *Session, pr *planRun) *applyRun {
	ar := &applyRun{
		o: o, sess: sess, planRun: pr,
		wantByKey: map[string]plan.Wanted{},
		entries:   map[string]inventory.Entry{},
	}
	for _, w := range pr.want {
		ar.wantByKey[plan.ReleaseKey(w.Name, w.Namespace)] = w
	}
	for _, e := range pr.inv {
		ar.entries[entryKey(e.Name, e.Namespace)] = e
	}
	return ar
}

func (ar *applyRun) run(ctx context.Context) error {
	o := ar.o
	debugf(o, ar.rd, "deployment=%s purge=%v timeout=%s", ar.sess.DeploymentID(), o.Purge, helmDefaults(o, ar.sess.Doc).Timeout)
	if err := refusedBlocking(ar.res); err != nil {
		return err
	}
	// Pull secrets follow the registry credentials, which the release plan does
	// not see, so they are synced on every apply.
	base := []engine.Step{ar.stepEnsureNamespaces(), ar.stepPullSecrets()}
	if !ar.res.HasChange {
		if err := engine.Run(ctx, base); err != nil {
			return err
		}
		fmt.Fprintln(o.out(), "No changes.")
		return nil
	}
	if err := ar.confirm(ctx); err != nil {
		return err
	}
	steps := base
	releaseSteps, err := ar.releaseSteps()
	if err != nil {
		return err
	}
	steps = append(steps, releaseSteps...)
	steps = append(steps, ar.stepRemovals())
	return engine.Run(ctx, steps)
}

func (ar *applyRun) confirm(ctx context.Context) error {
	removals := removalRows(ar.res)
	if hasNonRemovalChanges(ar.res) {
		if err := engine.ConfirmDeployApply(ar.o.Yes); err != nil {
			return err
		}
	}
	if len(removals) == 0 {
		return nil
	}
	if err := engine.ConfirmDeployRemovals(ar.o.Yes); err != nil {
		return err
	}
	if ar.o.Purge {
		return confirmPurgeRemovals(ctx, ar, removals)
	}
	return nil
}

func (ar *applyRun) writeInventory(ctx context.Context) error {
	return inventory.Write(ctx, ar.sess.Client, ar.sess.DeploymentID(), inventoryFromMap(ar.entries))
}

func (ar *applyRun) stepEnsureNamespaces() engine.Step {
	return stepFunc(func(ctx context.Context) error {
		return ensureNamespaces(ctx, ar.sess.Client, ar.sess.Doc)
	})
}

func (ar *applyRun) stepPullSecrets() engine.Step {
	return stepFunc(func(ctx context.Context) error {
		cred, err := ocireg.LoadAuth(ar.regCfg)
		if err != nil {
			return err
		}
		return ensurePullSecrets(ctx, ar.sess.Client, ar.sess.Doc, cred)
	})
}

// releaseWork is the state one release carries through its apply steps.
type releaseWork struct {
	rel    Release
	want   plan.Wanted
	values map[string]any
	ref    string
}

// releaseSteps returns resolve → crds → install → record steps for every
// release the plan changes, in dependency order.
func (ar *applyRun) releaseSteps() ([]engine.Step, error) {
	order, err := InstallOrder(ar.sess.Doc.Releases)
	if err != nil {
		return nil, err
	}
	byName := map[string]Release{}
	for _, r := range ar.sess.Doc.Releases {
		byName[r.Name] = r
	}
	rowByKey := map[string]plan.Row{}
	for _, row := range ar.res.Rows {
		rowByKey[plan.ReleaseKey(row.Name, row.Namespace)] = row
	}
	var steps []engine.Step
	for _, name := range order {
		rel := byName[name]
		row, ok := rowByKey[plan.ReleaseKey(rel.Name, rel.Namespace)]
		if !ok || !appliesChange(row.Action) {
			continue
		}
		if ar.o.Only != "" && !ar.closure[name] {
			continue
		}
		rw := &releaseWork{rel: rel}
		steps = append(steps,
			stepFunc(func(context.Context) error { return ar.resolveRelease(rw) }),
			stepFunc(func(ctx context.Context) error { return ar.upgradeCRDs(ctx, rw) }),
			stepFunc(func(ctx context.Context) error { return ar.installRelease(ctx, rw) }),
			stepFunc(func(ctx context.Context) error { return ar.recordRelease(ctx, rw) }),
		)
	}
	return steps, nil
}

func appliesChange(a plan.Action) bool {
	switch a {
	case plan.ActionInstall, plan.ActionUpgrade, plan.ActionAdopt:
		return true
	default:
		return false
	}
}

func (ar *applyRun) resolveRelease(rw *releaseWork) error {
	w, ok := ar.wantByKey[plan.ReleaseKey(rw.rel.Name, rw.rel.Namespace)]
	if !ok {
		return fmt.Errorf("release %q missing from wanted set", rw.rel.Name)
	}
	values, ok := ar.resolved[rw.rel.Name]
	if !ok {
		return fmt.Errorf("release %q missing resolved values", rw.rel.Name)
	}
	ref, err := ocireg.ChartRef(ar.regCfg, rw.rel.Chart)
	if err != nil {
		return err
	}
	rw.want, rw.values, rw.ref = w, values, ref
	return nil
}

func (ar *applyRun) upgradeCRDs(ctx context.Context, rw *releaseWork) error {
	if rw.rel.CRDs != "upgrade" {
		return nil
	}
	chartBytes, err := ar.helm.FetchChartArchive(ctx, rw.ref, rw.rel.Version)
	if err != nil {
		return err
	}
	return crds.Upgrade(ctx, ar.sess.Kube, rw.rel.CRDs, chartBytes)
}

func (ar *applyRun) installRelease(ctx context.Context, rw *releaseWork) error {
	return ar.helm.InstallOrUpgrade(ctx, helmclient.ReleaseRequest{
		Name: rw.rel.Name, Namespace: rw.rel.Namespace, Chart: rw.rel.Chart, Version: rw.rel.Version,
	}, rw.ref, rw.values)
}

func (ar *applyRun) recordRelease(ctx context.Context, rw *releaseWork) error {
	rel := rw.rel
	ar.entries[entryKey(rel.Name, rel.Namespace)] = inventory.Entry{
		Name: rel.Name, Namespace: rel.Namespace, Chart: rel.Chart, Version: rel.Version,
		ChartDigest: rw.want.ChartDigest, ValuesDigest: rw.want.FileValuesDigest,
		Requires: rel.Requires, Protect: rel.Protect, UpdatedAt: time.Now().UTC(),
	}
	return ar.writeInventory(ctx)
}

func (ar *applyRun) stepRemovals() engine.Step {
	return stepFunc(func(ctx context.Context) error {
		for _, row := range removalRows(ar.res) {
			if err := ar.removeRelease(ctx, row); err != nil {
				return err
			}
		}
		return nil
	})
}

func (ar *applyRun) removeRelease(ctx context.Context, row plan.Row) error {
	manifest, err := ar.helm.ReleaseManifest(ctx, row.Name, row.Namespace)
	if err != nil {
		return err
	}
	if err := ar.helm.Uninstall(ctx, row.Name, row.Namespace); err != nil {
		return err
	}
	delete(ar.entries, entryKey(row.Name, row.Namespace))
	if err := ar.writeInventory(ctx); err != nil {
		return err
	}
	opt := purgeRemovalOpts{ReleaseName: row.Name, ReleaseNS: row.Namespace, Manifest: manifest}
	return reportOrPurgeRemovalData(ctx, ar.sess.Client, ar.sess.Doc, ar.entries, opt, ar.o.Purge, ar.o.out())
}

func hasNonRemovalChanges(res plan.Result) bool {
	for _, row := range res.Rows {
		switch row.Action {
		case plan.ActionInstall, plan.ActionUpgrade, plan.ActionAdopt, plan.ActionStuck:
			return true
		}
	}
	return false
}

func removalRows(res plan.Result) []plan.Row {
	var out []plan.Row
	for _, row := range res.Removals {
		if row.Action == plan.ActionRemove {
			out = append(out, row)
		}
	}
	return out
}

type stepFunc func(ctx context.Context) error

func (f stepFunc) Run(ctx context.Context) error { return f(ctx) }
