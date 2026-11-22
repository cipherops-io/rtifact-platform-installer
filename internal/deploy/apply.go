package deploy

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"
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
}

func buildWanted(opts wantedBuildOpts) ([]plan.Wanted, error) {
	var out []plan.Wanted
	for _, rel := range opts.doc.Releases {
		merged, err := MergeReleaseValues(opts.deployPath, rel)
		if err != nil {
			return nil, err
		}
		valuesDigest, err := ValuesDigest(merged)
		if err != nil {
			return nil, err
		}
		chartDigest, err := ocireg.ChartDigest(opts.regClient, opts.reg, rel.Chart, rel.Version)
		if err != nil {
			chartDigest = ""
		}
		out = append(out, plan.Wanted{
			Name: rel.Name, Namespace: rel.Namespace, Chart: rel.Chart,
			Version: rel.Version, ChartDigest: chartDigest, ValuesDigest: valuesDigest,
			Requires: rel.Requires, Protect: rel.Protect,
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
}

func entryKey(name, ns string) string { return ns + "/" + name }

// Apply installs or upgrades releases.
func Apply(ctx context.Context, o RunOptions) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	ar, err := prepareApplyRun(ctx, o)
	if err != nil {
		return AsInterrupted(err)
	}
	printBanner(o.out(), ar.doc, ar.disc, ar.kcfg)

	if !ar.res.HasChange {
		fmt.Fprintln(o.out(), "No changes.")
		return nil
	}
	if err := refusedBlocking(ar.res); err != nil {
		return AsInterrupted(err)
	}
	if err := engine.ConfirmDeployApply(o.Yes, len(removalRows(ar.res)) > 0); err != nil {
		return AsInterrupted(err)
	}
	if err := ensureSystemNS(ctx, ar.kcfg); err != nil {
		return AsInterrupted(err)
	}
	if err := ar.acquireLock(); err != nil {
		return AsInterrupted(err)
	}
	defer func() {
		if ar.lockHold == nil {
			return
		}
		releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = ar.lockHold.Release(releaseCtx, ar.kcfg, ar.deployID)
	}()

	return AsInterrupted(engine.Run(ctx, []engine.Step{
		ar.stepEnsureNamespaces(),
		ar.stepPullSecrets(),
		ar.stepReleases(),
		ar.stepRemovals(),
	}))
}

func prepareApplyRun(ctx context.Context, o RunOptions) (*applyRun, error) {
	sess, err := OpenSession(ctx, o)
	if err != nil {
		return nil, err
	}
	doc, disc, kcfg := sess.Doc, sess.Disc, sess.Kube
	regCfg := registryFromDoc(doc)
	cred, err := ocireg.LoadAuth(regCfg)
	if err != nil {
		return nil, err
	}
	reg, err := ocireg.NewClient(regCfg, cred)
	if err != nil {
		return nil, err
	}
	want, err := buildWanted(wantedBuildOpts{
		deployPath: o.FilePath, doc: doc, outputsDir: disc.OutputsPath,
		reg: regCfg, regClient: reg,
	})
	if err != nil {
		return nil, err
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
		return nil, err
	}
	observed, err := cs.Observe(ctx, helmclient.ObserveOptions{Wanted: want})
	if err != nil {
		return nil, err
	}
	inv, err := inventory.Read(ctx, kcfg, DeploymentID(doc))
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
	printPlan(o.out(), res)

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
	}, nil
}

func (ar *applyRun) acquireLock() error {
	h, err := lock.Acquire(ar.ctx, ar.kcfg, ar.deployID, lock.HolderIdentity(newRunID()), ar.o.ForceUnlock)
	if err != nil {
		return err
	}
	ar.lockHold = h
	return nil
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
		rowByName := map[string]plan.Row{}
		for _, row := range ar.res.Rows {
			rowByName[row.Name] = row
		}
		for _, name := range order {
			row, ok := rowByName[name]
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
			rel := ar.relByName[name]
			w := wantByName(ar.want, name)
			if w == nil {
				return fmt.Errorf("release %q missing from wanted set", name)
			}
			merged, err := MergeReleaseValues(ar.o.FilePath, rel)
			if err != nil {
				return err
			}
			rd := redact.New()
			resolved, err := ResolveReleaseValuesFrom(merged, rd, ar.disc.OutputsPath)
			if err != nil {
				return err
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
				ChartDigest: chartDigest, ValuesDigest: w.ValuesDigest,
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
			if err := ar.helm.Uninstall(ctx, row.Name, ns); err != nil {
				return err
			}
			delete(ar.entries, entryKey(row.Name, row.Namespace))
			if err := ar.writeInventory(); err != nil {
				return err
			}
		}
		return nil
	})
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