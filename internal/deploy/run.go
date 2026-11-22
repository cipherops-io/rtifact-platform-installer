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
	sess, err := OpenSession(ctx, o)
	if err != nil {
		return 1, err
	}
	doc, disc, kcfg := sess.Doc, sess.Disc, sess.Kube
	printBanner(o.out(), doc, disc, kcfg)
	regCfg := registryFromDoc(doc)
	cred, err := ocireg.LoadAuth(regCfg)
	if err != nil {
		return 1, err
	}
	reg, err := ocireg.NewClient(regCfg, cred)
	if err != nil {
		return 1, err
	}
	cs, err := helmclient.New(helmclient.Options{
		Kube: kcfg, Deployment: DeploymentID(doc),
		Defaults: helmclient.Defaults{
			Timeout: doc.Defaults.Timeout, Wait: doc.Defaults.Wait,
			RollbackOnFailure: doc.Defaults.RollbackOnFailure, HistoryMax: doc.Defaults.HistoryMax,
		},
		Registry: helmclient.RegistryOpts{PlainHTTP: doc.Registry.PlainHTTP},
	})
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
	observed, err := cs.Observe(ctx, helmclient.ObserveOptions{Wanted: want})
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
		FilePath: o.FilePath,
	}, plan.Options{
		AllowDestroy: o.AllowDestroy, AllowMassRemoval: o.AllowMassRemoval,
		AllowDowngrade: o.AllowDowngrade, AllowAdopt: o.AllowAdopt, Purge: o.Purge,
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

func printPlan(w io.Writer, res plan.Result) {
	removalsShown := false
	for _, row := range res.Rows {
		note := ""
		if row.Note != "" {
			note = "  [" + row.Note + "]"
		}
		switch row.Action {
		case plan.ActionUnchanged:
			fmt.Fprintf(w, "  %s/%s: unchanged (%s)\n", row.Namespace, row.Name, row.ToVersion)
		case plan.ActionInstall:
			fmt.Fprintf(w, "  %s/%s: install %s%s\n", row.Namespace, row.Name, row.ToVersion, note)
		case plan.ActionUpgrade:
			fmt.Fprintf(w, "  %s/%s: upgrade %s -> %s%s\n", row.Namespace, row.Name, row.FromVersion, row.ToVersion, note)
		case plan.ActionRemove:
			if !removalsShown {
				fmt.Fprint(w, "\nRemovals:\n")
				removalsShown = true
			}
			fmt.Fprintf(w, "  %s/%s: remove%s\n", row.Namespace, row.Name, note)
		case plan.ActionRefused, plan.ActionStuck:
			fmt.Fprintf(w, "  %s/%s: %s %s\n", row.Namespace, row.Name, row.Action, row.Note)
		default:
			fmt.Fprintf(w, "  %s/%s: %s%s\n", row.Namespace, row.Name, row.Action, note)
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

// MaskedPrint applies env masking then redaction for user output.
func MaskedPrint(rd *redact.Redactor, s string) string {
	s = MaskEnvPlaceholders(s)
	return rd.Text(s)
}