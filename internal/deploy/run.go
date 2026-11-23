package deploy

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/ocireg"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/redact"
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
	JSONOutput        bool
	// NeedsRegistry is true for plan/apply (OCI chart digests); status, destroy, and rollback leave it false.
	NeedsRegistry     bool
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


// Plan prints pending changes without mutating the cluster.
func Plan(ctx context.Context, o RunOptions) (int, error) {
	sess, err := OpenSession(ctx, o)
	if err != nil {
		return 1, err
	}
	doc, disc, kcfg := sess.Doc, sess.Disc, sess.Kube
	PrintBanner(o.err(), doc, disc, kcfg, sess.Identity)
	rd := redact.New()
	resolved, err := ResolveAllReleases(o.FilePath, doc, disc.OutputsPath, rd)
	if err != nil {
		return 1, err
	}
	debugf(o, "registry host=%s plain_http=%v", doc.Registry.Host, doc.Registry.PlainHTTP)
	o.NeedsRegistry = true
	regCfg := registryFromDoc(doc)
	reg, err := registryClient(ctx, doc, true)
	if err != nil {
		return 1, err
	}
	cs, err := newHelmClient(kcfg, doc, reg, o)
	if err != nil {
		return 1, err
	}
	want, err := buildWanted(wantedBuildOpts{
		deployPath: o.FilePath, doc: doc, outputsDir: disc.OutputsPath,
		reg: regCfg, regClient: reg,
	})
	if err != nil {
		return 1, err
	}
	inv, err := inventory.Read(ctx, kcfg, DeploymentID(doc))
	if err != nil {
		return 1, err
	}
	observed, err := cs.Observe(ctx, helmclient.ObserveOptions{Wanted: observeTargets(want, inv)})
	if err != nil {
		return 1, err
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
		return 1, err
	}
	if err := printPlanManifestDiffs(ctx, planDiffOpts{
		out: o.out(), stderr: o.err(), debug: o.Debug,
		cs: cs, doc: doc, res: res, regCfg: regCfg,
		resolved: resolved, rd: rd,
	}); err != nil {
		return 1, err
	}
	if refusedBlocking(res) != nil {
		return 1, refusedBlocking(res)
	}
	if o.DetailedExitCode {
		if res.HasChange {
			return 2, nil
		}
		return 0, nil
	}
	return 0, nil
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

func refusedBlocking(res plan.Result) error {
	for _, row := range res.Rows {
		if row.Action == plan.ActionRefused {
			return fmt.Errorf("%s/%s: %s", row.Namespace, row.Name, row.Note)
		}
	}
	return nil
}

func newHelmClient(kcfg kube.Config, doc *Document, reg *registry.Client, o RunOptions) (*helmclient.Client, error) {
	deployment := DeploymentID(doc)
	return helmclient.New(helmclient.Options{
		Kube: kcfg, Deployment: deployment, RegistryClient: reg,
		Defaults: helmDefaults(o, doc),
		Registry: helmclient.RegistryOpts{PlainHTTP: doc.Registry.PlainHTTP},
	})
}

func registryClient(ctx context.Context, doc *Document, needs bool) (*registry.Client, error) {
	_ = ctx
	if !needs {
		return nil, nil
	}
	regCfg := registryFromDoc(doc)
	cred, err := ocireg.LoadAuth(regCfg)
	if err != nil {
		return nil, err
	}
	return ocireg.NewClient(regCfg, cred)
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
	existing, err := cs.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		merged := mergeNamespaceLabels(existing.Labels, labels)
		if labelsEqual(merged, existing.Labels) {
			return nil
		}
		existing.Labels = merged
		_, err = cs.CoreV1().Namespaces().Update(ctx, existing, metav1.UpdateOptions{})
		return err
	}
	if !apierrors.IsNotFound(err) {
		return err
	}
	createLabels := mergeNamespaceLabels(nil, labels)
	createLabels[LabelNamespaceCreatedBy] = LabelNamespaceCreatedByValue
	obj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: name, Labels: createLabels},
	}
	_, err = cs.CoreV1().Namespaces().Create(ctx, obj, metav1.CreateOptions{})
	return err
}

func mergeNamespaceLabels(base, extra map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

func labelsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
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

// MaskedPrint applies env masking then redaction for user output.
func MaskedPrint(rd *redact.Redactor, s string) string {
	s = MaskEnvPlaceholders(s)
	return rd.Text(s)
}