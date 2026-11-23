package deploy

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/crds"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/lock"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/ocireg"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/redact"
	"github.com/cipherops-io/rtifact-platform-installer/internal/engine"
	"helm.sh/helm/v4/pkg/registry"
)

type wantedBuildOpts struct {
	deployPath string
	doc        *Document
	outputsDir string
	reg        ocireg.Registry
	regClient  *registry.Client
	rd         *redact.Redactor
}

func buildWanted(opts wantedBuildOpts) ([]plan.Wanted, error) {
	var out []plan.Wanted
	for _, rel := range opts.doc.Releases {
		merged, err := MergeReleaseValues(opts.deployPath, rel)
		if err != nil {
			return nil, err
		}
		fileDigest, err := ValuesDigest(merged)
		if err != nil {
			return nil, err
		}
		resolved, err := ResolveReleaseValuesFrom(merged, opts.rd, opts.outputsDir)
		if err != nil {
			return nil, err
		}
		valuesDigest, err := ValuesDigest(resolved)
		if err != nil {
			return nil, err
		}
		chartDigest, err := ocireg.ChartDigest(opts.regClient, opts.reg, rel.Chart, rel.Version)
		if err != nil {
			return nil, err
		}
		out = append(out, plan.Wanted{
			Name: rel.Name, Namespace: rel.Namespace, Chart: rel.Chart,
			Version: rel.Version, ChartDigest: chartDigest, ValuesDigest: valuesDigest,
			FileValuesDigest: fileDigest,
			Requires:         rel.Requires, Protect: rel.Protect,
		})
	}
	return out, nil
}

type applyRun struct {
	ctx      context.Context
	o        RunOptions
	doc      *Document
	disc     Discovery
	kcfg     kube.Config
	regCfg   ocireg.Registry
	reg      *registry.Client
	deployID string

	want     []plan.Wanted
	res      plan.Result
	inv      []inventory.Entry
	closure  map[string]bool
	helm     *helmclient.Client
	lockHold *lock.Handle

	relByName map[string]Release
	entries   map[string]inventory.Entry
	identity  IdentityInfo
	rd        *redact.Redactor
	resolved  map[string]map[string]any
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
	deployID := DeploymentID(sess.Doc)
	if err := ensureSystemNS(ctx, sess.Kube); err != nil {
		return AsInterrupted(err)
	}
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	h, err := lock.Acquire(runCtx, sess.Kube, deployID, lock.HolderIdentity(newRunID()), o.ForceUnlock)
	if err != nil {
		return AsInterrupted(err)
	}
	h.OnRenewFailure(func(err error) {
		fmt.Fprintf(o.err(), "deploy lock renewal failed; stopping apply: %v\n", err)
		cancel()
	})
	defer func() {
		releaseCtx, rcancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer rcancel()
		_ = h.Release(releaseCtx, sess.Kube, deployID)
	}()

	ar, err := prepareApplyRun(runCtx, o, sess)
	if err != nil {
		return AsInterrupted(RedactError(ActiveRedactor(), err))
	}
	ar.lockHold = h
	return AsInterrupted(RedactError(ar.rd, ar.run(runCtx)))
}

func (ar *applyRun) run(ctx context.Context) error {
	o := ar.o
	debugf(o, ar.rd, "deployment=%s purge=%v timeout=%s", ar.deployID, o.Purge, helmDefaults(o, ar.doc).Timeout)
	if !ar.res.HasChange {
		fmt.Fprintln(o.out(), "No changes.")
		return nil
	}
	if err := refusedBlocking(ar.res); err != nil {
		return err
	}
	removals := removalRows(ar.res)
	if hasNonRemovalChanges(ar.res) {
		if err := engine.ConfirmDeployApply(o.Yes); err != nil {
			return err
		}
	}
	if len(removals) > 0 {
		if err := engine.ConfirmDeployRemovals(o.Yes); err != nil {
			return err
		}
	}
	if o.Purge && len(removals) > 0 {
		if err := confirmPurgeRemovals(ctx, ar, removals); err != nil {
			return err
		}
	}
	return engine.Run(ctx, []engine.Step{
		ar.stepEnsureNamespaces(),
		ar.stepPullSecrets(),
		ar.stepReleases(),
		ar.stepRemovals(),
	})
}

func prepareApplyRun(ctx context.Context, o RunOptions, sess *Session) (*applyRun, error) {
	doc, disc, kcfg := sess.Doc, sess.Disc, sess.Kube
	rd := redact.New()
	resolved, err := ResolveAllReleases(o.FilePath, doc, disc.OutputsPath, rd)
	if err != nil {
		return nil, err
	}
	SetActiveRedactor(rd)
	o.withRedactedOutput(rd)
	o.NeedsRegistry = true
	regCfg := registryFromDoc(doc)
	reg, err := registryClient(ctx, doc, true)
	if err != nil {
		return nil, err
	}
	want, err := buildWanted(wantedBuildOpts{
		deployPath: o.FilePath, doc: doc, outputsDir: disc.OutputsPath,
		reg: regCfg, regClient: reg, rd: rd,
	})
	if err != nil {
		return nil, err
	}
	cs, err := newHelmClient(kcfg, doc, reg, o)
	if err != nil {
		return nil, err
	}
	inv, err := inventory.Read(ctx, kcfg, DeploymentID(doc))
	if err != nil {
		return nil, err
	}
	observed, err := cs.Observe(ctx, helmclient.ObserveOptions{Wanted: observeTargets(want, inv)})
	if err != nil {
		return nil, err
	}
	closure := onlyClosure(doc, o.Only)
	res := plan.Build(plan.Input{
		Wanted: want, Observed: observed, Inventory: inv, Deployment: DeploymentID(doc),
		FilePath: o.FilePath,
	}, plan.Options{
		AllowDestroy: o.AllowDestroy, AllowMassRemoval: o.AllowMassRemoval,
		AllowDowngrade: o.AllowDowngrade, AllowAdopt: o.AllowAdopt, Purge: o.Purge,
		Only: o.Only, OnlyClosure: closure,
	})
	if err := PrintPlan(o.out(), rd, res, o.JSONOutput); err != nil {
		return nil, err
	}
	if err := printPlanManifestDiffs(ctx, planDiffOpts{
		out: o.out(), stderr: o.err(), debug: o.Debug,
		cs: cs, doc: doc, res: res, regCfg: regCfg,
		resolved: resolved, rd: rd,
	}); err != nil {
		return nil, err
	}

	relByName := map[string]Release{}
	for _, r := range doc.Releases {
		relByName[r.Name] = r
	}
	entries := map[string]inventory.Entry{}
	for _, e := range inv {
		entries[entryKey(e.Name, e.Namespace)] = e
	}

	return &applyRun{
		ctx: ctx, o: o, doc: doc, disc: disc, kcfg: kcfg, regCfg: regCfg, reg: reg,
		deployID: DeploymentID(doc), want: want, res: res, inv: inv, closure: closure,
		helm: cs, relByName: relByName, entries: entries,
		identity: sess.Identity, rd: rd, resolved: resolved,
	}, nil
}

func (ar *applyRun) writeInventory() error {
	return inventory.Write(ar.ctx, ar.kcfg, ar.deployID, inventoryEntries(ar.entries))
}

func inventoryEntries(m map[string]inventory.Entry) []inventory.Entry {
	out := make([]inventory.Entry, 0, len(m))
	for _, e := range m {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Namespace != out[j].Namespace {
			return out[i].Namespace < out[j].Namespace
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func (ar *applyRun) stepEnsureNamespaces() engine.Step {
	return stepFunc(func(ctx context.Context) error {
		return ensureNamespaces(ctx, ar.kcfg, ar.doc)
	})
}

func (ar *applyRun) stepPullSecrets() engine.Step {
	return stepFunc(func(ctx context.Context) error {
		cred, err := ocireg.LoadAuth(ar.regCfg)
		if err != nil {
			return err
		}
		return ensurePullSecrets(ctx, ar.kcfg, ar.doc, cred)
	})
}

func (ar *applyRun) stepReleases() engine.Step {
	return stepFunc(func(ctx context.Context) error {
		order, err := InstallOrder(ar.doc.Releases)
		if err != nil {
			return err
		}
		rowByKey := map[string]plan.Row{}
		for _, row := range ar.res.Rows {
			rowByKey[plan.ReleaseKey(row.Name, row.Namespace)] = row
		}
		for _, name := range order {
			rel := ar.relByName[name]
			row, ok := rowByKey[plan.ReleaseKey(rel.Name, rel.Namespace)]
			if !ok {
				continue
			}
			switch row.Action {
			case plan.ActionUnchanged, plan.ActionRefused, plan.ActionStuck:
				continue
			case plan.ActionInstall, plan.ActionUpgrade, plan.ActionAdopt:
			default:
				continue
			}
			if ar.o.Only != "" && !ar.closure[name] {
				continue
			}
			w := wantByName(ar.want, name)
			if w == nil {
				return fmt.Errorf("release %q missing from wanted set", name)
			}
			resolved, ok := ar.resolved[name]
			if !ok {
				return fmt.Errorf("release %q missing resolved values", name)
			}
			ref, err := ocireg.ChartRef(ar.regCfg, rel.Chart)
			if err != nil {
				return err
			}
			chartDigest, err := ocireg.ChartDigest(ar.reg, ar.regCfg, rel.Chart, rel.Version)
			if err != nil {
				return err
			}
			if chartDigest == "" {
				chartDigest = w.ChartDigest
			}
			if rel.CRDs == "upgrade" {
				chartBytes, err := ar.helm.FetchChartArchive(ctx, ref, rel.Version)
				if err != nil {
					return err
				}
				if err := crds.Upgrade(ctx, ar.kcfg, rel.CRDs, chartBytes); err != nil {
					return err
				}
			}
			if err := ar.helm.InstallOrUpgrade(ctx, helmclient.ReleaseRequest{
				Name: rel.Name, Namespace: rel.Namespace, Chart: rel.Chart, Version: rel.Version,
			}, ref, resolved, w.ValuesDigest); err != nil {
				return err
			}
			applied := inventory.Entry{
				Name: rel.Name, Namespace: rel.Namespace, Chart: rel.Chart, Version: rel.Version,
				ChartDigest: chartDigest, ValuesDigest: w.FileValuesDigest,
				Requires: rel.Requires, Protect: rel.Protect, UpdatedAt: time.Now().UTC(),
			}
			ar.entries[entryKey(rel.Name, rel.Namespace)] = applied
			if err := ar.writeInventory(); err != nil {
				return err
			}
		}
		return nil
	})
}

func (ar *applyRun) stepRemovals() engine.Step {
	return stepFunc(func(ctx context.Context) error {
		for _, row := range removalRows(ar.res) {
			ns := row.Namespace
			if e, ok := ar.entries[entryKey(row.Name, row.Namespace)]; ok && e.Namespace != "" {
				ns = e.Namespace
			} else {
				for _, invE := range ar.inv {
					if invE.Name == row.Name && invE.Namespace == row.Namespace {
						ns = invE.Namespace
						break
					}
				}
			}
			manifest, err := fetchReleaseManifest(ctx, ar.helm, row.Name, ns)
			if err != nil {
				return err
			}
			if err := ar.helm.Uninstall(ctx, row.Name, ns); err != nil {
				return err
			}
			delete(ar.entries, entryKey(row.Name, row.Namespace))
			if err := ar.writeInventory(); err != nil {
				return err
			}
			opt := purgeRemovalOpts{ReleaseName: row.Name, ReleaseNS: ns, Manifest: manifest}
			if err := reportOrPurgeRemovalData(ctx, ar.kcfg, ar.doc, ar.entries, opt, ar.o.Purge, ar.o.out()); err != nil {
				return err
			}
		}
		return nil
	})
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