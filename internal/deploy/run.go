package deploy

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/lock"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/ocireg"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/redact"
	"github.com/cipherops-io/rtifact-platform-installer/internal/engine"
	"helm.sh/helm/v4/pkg/registry"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RunOptions common CLI flags.
type RunOptions struct {
	FilePath          string
	KubeconfigCLI     string
	Yes               bool
	Only              string
	DetailedExitCode  bool
	Timeout           time.Duration
	AllowDestroy      bool
	AllowMassRemoval  bool
	AllowDowngrade    bool
	AllowAdopt        bool
	ForceUnlock       bool
	Purge             bool
	Debug             bool
	Stdout            io.Writer
	Stderr            io.Writer
}

func (o *RunOptions) out() io.Writer {
	if o.Stdout != nil {
		return o.Stdout
	}
	return os.Stdout
}

func (o *RunOptions) err() io.Writer {
	if o.Stderr != nil {
		return o.Stderr
	}
	return os.Stderr
}

func loadRuntime(o RunOptions) (*Document, Discovery, kube.Config, error) {
	doc, err := LoadFile(o.FilePath)
	if err != nil {
		return nil, Discovery{}, kube.Config{}, err
	}
	disc, err := Discover(DiscoverOpts{
		FilePath: o.FilePath, Doc: doc, KubeconfigCLI: o.KubeconfigCLI,
	})
	if err != nil {
		return nil, disc, kube.Config{}, err
	}
	kcfg := kube.Config{
		Kubeconfig: disc.KubeconfigPath,
		Context:    doc.Cluster.Context,
	}
	if err := CheckIdentity(kcfg, doc, disc); err != nil {
		return nil, disc, kcfg, err
	}
	return doc, disc, kcfg, nil
}

func printBanner(w io.Writer, doc *Document, disc Discovery, kcfg kube.Config) {
	server, _ := kube.ServerURL(kcfg)
	fmt.Fprintf(w, "Cluster:    %s  %s\n", doc.NamePrefix, server)
	fmt.Fprintf(w, "Kubeconfig: %s (%s)\n", disc.KubeconfigPath, disc.KubeconfigFrom)
	if disc.OutputsPath != "" {
		fmt.Fprintf(w, "Outputs:    %s (%s)\n", disc.OutputsPath, disc.OutputsFrom)
	}
}

// Plan prints pending changes without mutating the cluster.
func Plan(ctx context.Context, o RunOptions) (int, error) {
	doc, disc, kcfg, err := loadRuntime(o)
	if err != nil {
		return 1, err
	}
	printBanner(o.out(), doc, disc, kcfg)
	regCfg := registryFromDoc(doc)
	cred, err := ocireg.LoadAuth(regCfg)
	if err != nil {
		return 1, err
	}
	if _, err := ocireg.NewClient(regCfg, cred); err != nil {
		return 1, err
	}
	cs, err := helmclient.New(helmclient.Options{
		Kube: kcfg, Deployment: DeploymentID(doc),
		Defaults: helmclient.Defaults{
			Timeout: doc.Defaults.Timeout, Wait: doc.Defaults.Wait,
			RollbackOnFailure: doc.Defaults.RollbackOnFailure, HistoryMax: doc.Defaults.HistoryMax,
		},
		Registry:       helmclient.RegistryOpts{PlainHTTP: doc.Registry.PlainHTTP},
		RegistryClient: mustReg(regCfg, cred),
	})
	if err != nil {
		return 1, err
	}
	want, err := buildWanted(o.FilePath, doc)
	if err != nil {
		return 1, err
	}
	observed, err := cs.Observe(ctx)
	if err != nil {
		return 1, err
	}
	inv, err := inventory.Read(ctx, kcfg, DeploymentID(doc))
	if err != nil {
		return 1, err
	}
	closure := onlyClosure(doc, o.Only)
	res := plan.Build(plan.Input{
		Wanted: want, Observed: observed, Inventory: inv, Deployment: DeploymentID(doc),
	}, plan.Options{
		AllowDestroy: o.AllowDestroy, AllowMassRemoval: o.AllowMassRemoval,
		AllowDowngrade: o.AllowDowngrade, AllowAdopt: o.AllowAdopt,
		Only: o.Only, OnlyClosure: closure,
	})
	printPlan(o.out(), res)
	if o.DetailedExitCode {
		if res.HasChange {
			return 2, nil
		}
		return 0, nil
	}
	return 0, nil
}

// Apply installs or upgrades releases.
func Apply(ctx context.Context, o RunOptions) error {
	doc, disc, kcfg, err := loadRuntime(o)
	if err != nil {
		return err
	}
	printBanner(o.out(), doc, disc, kcfg)
	regCfg := registryFromDoc(doc)
	cred, err := ocireg.LoadAuth(regCfg)
	if err != nil {
		return err
	}
	reg, err := ocireg.NewClient(regCfg, cred)
	if err != nil {
		return err
	}
	h, err := lock.Acquire(ctx, kcfg, DeploymentID(doc), hostname(), o.ForceUnlock)
	if err != nil {
		return err
	}
	defer h.Release(ctx, kcfg, DeploymentID(doc))

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
	want, err := buildWanted(o.FilePath, doc)
	if err != nil {
		return err
	}
	observed, err := cs.Observe(ctx)
	if err != nil {
		return err
	}
	inv, err := inventory.Read(ctx, kcfg, DeploymentID(doc))
	if err != nil {
		return err
	}
	closure := onlyClosure(doc, o.Only)
	res := plan.Build(plan.Input{
		Wanted: want, Observed: observed, Inventory: inv, Deployment: DeploymentID(doc),
	}, plan.Options{
		AllowDestroy: o.AllowDestroy, AllowMassRemoval: o.AllowMassRemoval,
		AllowDowngrade: o.AllowDowngrade, AllowAdopt: o.AllowAdopt,
		Only: o.Only, OnlyClosure: closure,
	})
	printPlan(o.out(), res)
	if !res.HasChange {
		fmt.Fprintln(o.out(), "No changes.")
		return nil
	}
	if err := engine.ConfirmDeployApply(o.Yes, len(res.Removals) > 0); err != nil {
		return err
	}
	if err := ensureSystemNS(ctx, kcfg); err != nil {
		return err
	}
	if err := ensureNamespaces(ctx, kcfg, doc); err != nil {
		return err
	}
	order, err := InstallOrder(doc.Releases)
	if err != nil {
		return err
	}
	relByName := map[string]Release{}
	for _, r := range doc.Releases {
		relByName[r.Name] = r
	}
	rowByName := map[string]plan.Row{}
	for _, row := range res.Rows {
		rowByName[row.Name] = row
	}
	var entries []inventory.Entry
	for _, name := range order {
		row, ok := rowByName[name]
		if !ok {
			continue
		}
		if row.Action == plan.ActionUnchanged || row.Action == plan.ActionRefused || row.Action == plan.ActionStuck {
			w := wantByName(want, name)
			if w != nil {
				entries = append(entries, inventory.Entry{
					Name: w.Name, Namespace: w.Namespace, Chart: w.Chart,
					Version: w.Version, ValuesDigest: w.ValuesDigest, Protect: w.Protect,
				})
			}
			continue
		}
		if row.Action != plan.ActionInstall && row.Action != plan.ActionUpgrade && row.Action != plan.ActionAdopt {
			continue
		}
		if o.Only != "" && !closure[name] {
			continue
		}
		rel := relByName[name]
		merged, err := MergeReleaseValues(o.FilePath, rel)
		if err != nil {
			return err
		}
		rd := redact.New()
		resolved, err := ResolveReleaseValues(merged, rd)
		if err != nil {
			return err
		}
		digest, err := ValuesDigest(resolved)
		if err != nil {
			return err
		}
		ref, err := ocireg.ChartRef(regCfg, rel.Chart)
		if err != nil {
			return err
		}
		if err := cs.InstallOrUpgrade(ctx, helmclient.ReleaseRequest{
			Name: rel.Name, Namespace: rel.Namespace, Chart: rel.Chart, Version: rel.Version,
		}, ref, resolved, digest); err != nil {
			return err
		}
		entries = append(entries, inventory.Entry{
			Name: rel.Name, Namespace: rel.Namespace, Chart: rel.Chart,
			Version: rel.Version, ValuesDigest: digest, Protect: rel.Protect,
		})
	}
	// Keep unchanged releases in inventory.
	for _, w := range want {
		if findEntry(entries, w.Name) {
			continue
		}
		entries = append(entries, inventory.Entry{
			Name: w.Name, Namespace: w.Namespace, Chart: w.Chart,
			Version: w.Version, ValuesDigest: w.ValuesDigest, Protect: w.Protect,
		})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return inventory.Write(ctx, kcfg, DeploymentID(doc), entries)
}

func buildWanted(path string, doc *Document) ([]plan.Wanted, error) {
	var out []plan.Wanted
	for _, rel := range doc.Releases {
		merged, err := MergeReleaseValues(path, rel)
		if err != nil {
			return nil, err
		}
		rd := redact.New()
		resolved, err := ResolveReleaseValues(merged, rd)
		if err != nil {
			return nil, err
		}
		digest, err := ValuesDigest(resolved)
		if err != nil {
			return nil, err
		}
		out = append(out, plan.Wanted{
			Name: rel.Name, Namespace: rel.Namespace, Chart: rel.Chart,
			Version: rel.Version, ValuesDigest: digest, Protect: rel.Protect,
		})
	}
	return out, nil
}

func onlyClosure(doc *Document, only string) map[string]bool {
	if only == "" {
		return nil
	}
	set := map[string]bool{only: true}
	for {
		changed := false
		for _, r := range doc.Releases {
			if !set[r.Name] {
				continue
			}
			for _, req := range r.Requires {
				if !set[req] {
					set[req] = true
					changed = true
				}
			}
		}
		if !changed {
			break
		}
	}
	return set
}

func printPlan(w io.Writer, res plan.Result) {
	for _, row := range res.Rows {
		switch row.Action {
		case plan.ActionUnchanged:
			fmt.Fprintf(w, "  %s/%s: unchanged (%s)\n", row.Namespace, row.Name, row.ToVersion)
		case plan.ActionInstall:
			fmt.Fprintf(w, "  %s/%s: install %s\n", row.Namespace, row.Name, row.ToVersion)
		case plan.ActionUpgrade:
			fmt.Fprintf(w, "  %s/%s: upgrade %s -> %s\n", row.Namespace, row.Name, row.FromVersion, row.ToVersion)
		case plan.ActionRemove:
			fmt.Fprintf(w, "\nRemovals:\n  %s/%s: remove\n", row.Namespace, row.Name)
		case plan.ActionRefused, plan.ActionStuck:
			fmt.Fprintf(w, "  %s/%s: %s %s\n", row.Namespace, row.Name, row.Action, row.Note)
		default:
			fmt.Fprintf(w, "  %s/%s: %s\n", row.Namespace, row.Name, row.Action)
		}
	}
}

func registryFromDoc(doc *Document) ocireg.Registry {
	r := ocireg.Registry{
		Host: doc.Registry.Host, ChartsPath: doc.Registry.ChartsPath,
		CAFile: doc.Registry.CAFile, PlainHTTP: doc.Registry.PlainHTTP,
	}
	if doc.Registry.Auth != nil {
		r.UsernameEnv = doc.Registry.Auth.UsernameEnv
		r.PasswordEnv = doc.Registry.Auth.PasswordEnv
	}
	return r
}

func mustReg(reg ocireg.Registry, cred ocireg.Credentials) *registry.Client {
	c, err := ocireg.NewClient(reg, cred)
	if err != nil {
		return nil
	}
	return c
}

func ensureSystemNS(ctx context.Context, kcfg kube.Config) error {
	return ensureNamespace(ctx, kcfg, inventory.SystemNS, nil)
}

func ensureNamespaces(ctx context.Context, kcfg kube.Config, doc *Document) error {
	for _, ns := range doc.Namespaces {
		if err := ensureNamespace(ctx, kcfg, ns.Name, ns.Labels); err != nil {
			return err
		}
	}
	return nil
}

func ensureNamespace(ctx context.Context, kcfg kube.Config, name string, labels map[string]string) error {
	cs, err := kube.Clientset(kcfg)
	if err != nil {
		return err
	}
	_, err = cs.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return err
	}
	obj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels},
	}
	_, err = cs.CoreV1().Namespaces().Create(ctx, obj, metav1.CreateOptions{})
	return err
}

func hostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}

func wantByName(want []plan.Wanted, name string) *plan.Wanted {
	for i := range want {
		if want[i].Name == name {
			return &want[i]
		}
	}
	return nil
}

func findEntry(entries []inventory.Entry, name string) bool {
	for _, e := range entries {
		if e.Name == name {
			return true
		}
	}
	return false
}

// MaskedPrint applies env masking then redaction for user output.
func MaskedPrint(rd *redact.Redactor, s string) string {
	s = MaskEnvPlaceholders(s)
	return rd.Text(s)
}