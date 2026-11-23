package deploy

import (
	"context"
	"fmt"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/ocireg"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/redact"
	"helm.sh/helm/v4/pkg/registry"
)

// planRun is the observed state and decisions shared by plan and apply.
type planRun struct {
	rd       *redact.Redactor
	resolved map[string]map[string]any
	regCfg   ocireg.Registry
	reg      *registry.Client
	helm     helmclient.Interface
	want     []plan.Wanted
	inv      []inventory.Entry
	closure  map[string]bool
	res      plan.Result
}

// buildPlan resolves values, observes the cluster and builds the plan. It binds
// the session redactor and redacts o's writers before anything is printed.
func buildPlan(ctx context.Context, sess *Session, o *RunOptions) (*planRun, error) {
	doc := sess.Doc
	pr := &planRun{rd: redact.New()}
	SetActiveRedactor(pr.rd)
	var err error
	pr.resolved, err = ResolveAllReleases(o.FilePath, doc, sess.Disc.OutputsPath, pr.rd)
	if err != nil {
		return nil, err
	}
	o.withRedactedOutput(pr.rd)
	debugf(*o, pr.rd, "registry host=%s plain_http=%v", doc.Registry.Host, doc.Registry.PlainHTTP)
	pr.regCfg = registryFromDoc(doc)
	if pr.reg, err = registryClient(pr.regCfg); err != nil {
		return nil, err
	}
	hc, err := sess.helm(*o, pr.reg)
	if err != nil {
		return nil, err
	}
	pr.helm = hc
	pr.want, err = buildWanted(wantedBuildOpts{
		deployPath: o.FilePath, doc: doc, outputsDir: sess.Disc.OutputsPath,
		reg: pr.regCfg, regClient: pr.reg, rd: pr.rd,
	})
	if err != nil {
		return nil, err
	}
	if pr.inv, err = inventory.Read(ctx, sess.Client, sess.DeploymentID()); err != nil {
		return nil, err
	}
	observed, err := pr.helm.Observe(ctx, helmclient.ObserveOptions{Wanted: observeTargets(pr.want, pr.inv)})
	if err != nil {
		return nil, err
	}
	opts := o.Guards.planOptions(doc)
	pr.closure = opts.OnlyClosure
	pr.res = plan.Build(plan.Input{
		Wanted: pr.want, Observed: observed, Inventory: pr.inv,
		Deployment: sess.DeploymentID(), FilePath: o.FilePath,
	}, opts)
	return pr, nil
}

// print writes the plan table (or JSON) and manifest diffs to o's writers.
func (pr *planRun) print(ctx context.Context, sess *Session, o RunOptions) error {
	if err := PrintPlan(o.out(), pr.rd, pr.res, o.JSONOutput); err != nil {
		return err
	}
	return printPlanManifestDiffs(ctx, planDiffOpts{
		out: o.out(), stderr: o.err(), debug: o.Debug,
		cs: pr.helm, doc: sess.Doc, res: pr.res, regCfg: pr.regCfg,
		resolved: pr.resolved, rd: pr.rd,
	})
}

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
		w, err := wantedRelease(opts, rel)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, nil
}

func wantedRelease(opts wantedBuildOpts, rel Release) (plan.Wanted, error) {
	merged, err := MergeReleaseValues(opts.deployPath, rel)
	if err != nil {
		return plan.Wanted{}, err
	}
	fileDigest, err := ValuesDigest(merged)
	if err != nil {
		return plan.Wanted{}, err
	}
	resolved, err := ResolveReleaseValuesFrom(merged, opts.rd, opts.outputsDir)
	if err != nil {
		return plan.Wanted{}, err
	}
	valuesDigest, err := ValuesDigest(resolved)
	if err != nil {
		return plan.Wanted{}, err
	}
	chartDigest, err := ocireg.ChartDigest(opts.regClient, opts.reg, rel.Chart, rel.Version)
	if err != nil {
		return plan.Wanted{}, fmt.Errorf("release %q: chart digest: %w", rel.Name, err)
	}
	return plan.Wanted{
		Name: rel.Name, Namespace: rel.Namespace, Chart: rel.Chart,
		Version: rel.Version, ChartDigest: chartDigest, ValuesDigest: valuesDigest,
		FileValuesDigest: fileDigest, Requires: rel.Requires, Protect: rel.Protect,
	}, nil
}
