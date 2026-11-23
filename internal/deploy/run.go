package deploy

import (
	"context"
	"fmt"
	"time"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/cluster"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/lock"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/ocireg"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/redact"
	"helm.sh/helm/v4/pkg/registry"
)

// Plan prints pending changes without mutating the cluster.
// Exit codes: 0 no changes (or no --detailed-exitcode), 1 refusal or error, 2 pending changes.
func Plan(ctx context.Context, o RunOptions) (int, error) {
	defer SetActiveRedactor(nil)
	sess, err := OpenSession(ctx, o)
	if err != nil {
		return 1, err
	}
	PrintBanner(o.err(), sess.Doc, sess.Disc, sess.Kube, sess.Identity)
	pr, err := buildPlan(ctx, sess, &o)
	if err != nil {
		return 1, RedactError(ActiveRedactor(), err)
	}
	if err := pr.print(ctx, sess, o); err != nil {
		return 1, RedactError(pr.rd, err)
	}
	if err := refusedBlocking(pr.res); err != nil {
		return 1, RedactError(pr.rd, err)
	}
	if o.DetailedExitCode && pr.res.HasChange {
		return 2, nil
	}
	return 0, nil
}

// withLock holds the deployment lock around fn. A failed lease renewal cancels
// fn's context so the run stops instead of continuing unlocked.
func withLock(ctx context.Context, sess *Session, o RunOptions, cmd string, fn func(context.Context) error) error {
	if err := ensureSystemNS(ctx, sess.Client); err != nil {
		return err
	}
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	h, err := lock.Acquire(runCtx, sess.Client, sess.DeploymentID(), lock.HolderIdentity(newRunID()), o.ForceUnlock)
	if err != nil {
		return err
	}
	h.OnRenewFailure(func(err error) {
		fmt.Fprintf(o.err(), "deploy lock renewal failed; stopping %s: %v\n", cmd, err)
		cancel()
	})
	defer func() {
		releaseCtx, rcancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer rcancel()
		_ = h.Release(releaseCtx, sess.Client)
	}()
	return fn(runCtx)
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

func registryClient(regCfg ocireg.Registry) (*registry.Client, error) {
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

func ensureSystemNS(ctx context.Context, cs cluster.Client) error {
	return cluster.EnsureNamespace(ctx, cs, inventory.SystemNS, nil)
}

func ensureNamespaces(ctx context.Context, cs cluster.Client, doc *Document) error {
	for _, ns := range doc.Namespaces {
		if err := cluster.EnsureNamespace(ctx, cs, ns.Name, ns.Labels); err != nil {
			return err
		}
	}
	return nil
}

// MaskedPrint applies env masking then redaction for user output.
func MaskedPrint(rd *redact.Redactor, s string) string {
	s = MaskEnvPlaceholders(s)
	return rd.Text(s)
}
