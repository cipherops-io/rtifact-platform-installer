package helmclient

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
	"helm.sh/helm/v4/pkg/action"
	"helm.sh/helm/v4/pkg/chart/v2/loader"
	"helm.sh/helm/v4/pkg/cli"
	helmkube "helm.sh/helm/v4/pkg/kube"
	release "helm.sh/helm/v4/pkg/release/v1"
	"helm.sh/helm/v4/pkg/registry"
)

// ReleaseRequest is one install/upgrade target.
type ReleaseRequest struct {
	Name, Namespace string
	Chart, Version  string
}

// Defaults from deploy file defaults block.
type Defaults struct {
	Timeout           string
	Wait              *bool
	RollbackOnFailure *bool
	HistoryMax        int
}

// RegistryOpts for plain HTTP pulls.
type RegistryOpts struct {
	PlainHTTP bool
}

// Client wraps Helm SDK install/upgrade/list for deploy.
type Client struct {
	cfg        *action.Configuration
	settings   *cli.EnvSettings
	reg        *registry.Client
	deployment string
	def        Defaults
	regOpts    RegistryOpts
	timeout    time.Duration
	wait       bool
	rollback   bool
	historyMax int
}

// Options for Helm client construction.
type Options struct {
	Kube           kube.Config
	RegistryClient *registry.Client
	Defaults       Defaults
	Registry       RegistryOpts
	Deployment     string
}

// New builds a Helm client from explicit kubeconfig (not ambient KUBECONFIG).
func New(opts Options) (*Client, error) {
	settings := cli.New()
	settings.KubeConfig = opts.Kube.Kubeconfig
	settings.KubeContext = opts.Kube.Context
	cfg := action.NewConfiguration()
	driver := os.Getenv("HELM_DRIVER")
	if driver == "" {
		driver = "secret"
	}
	if err := cfg.Init(settings.RESTClientGetter(), "", driver); err != nil {
		return nil, err
	}
	if opts.RegistryClient != nil {
		cfg.RegistryClient = opts.RegistryClient
	}
	wait := true
	if opts.Defaults.Wait != nil {
		wait = *opts.Defaults.Wait
	}
	rb := true
	if opts.Defaults.RollbackOnFailure != nil {
		rb = *opts.Defaults.RollbackOnFailure
	}
	tout, err := time.ParseDuration(opts.Defaults.Timeout)
	if err != nil {
		tout = 10 * time.Minute
	}
	return &Client{
		cfg:        cfg,
		settings:   settings,
		reg:        cfg.RegistryClient,
		deployment: opts.Deployment,
		def:        opts.Defaults,
		regOpts:    opts.Registry,
		timeout:    tout,
		wait:       wait,
		rollback:   rb,
		historyMax: opts.Defaults.HistoryMax,
	}, nil
}

// Observe lists Helm releases labeled as managed by this deployment.
func (c *Client) Observe(ctx context.Context) ([]plan.Observed, error) {
	_ = ctx
	lst := action.NewList(c.cfg)
	lst.AllNamespaces = true
	lst.StateMask = action.ListAll
	rels, err := lst.Run()
	if err != nil {
		return nil, err
	}
	var out []plan.Observed
	for _, ri := range rels {
		rel, ok := ri.(*release.Release)
		if !ok {
			continue
		}
		if rel.Labels == nil || rel.Labels[LabelDeployment] != c.deployment {
			continue
		}
		chartName := rel.Chart.Metadata.Name
		if rel.Labels[LabelChart] != "" {
			chartName = rel.Labels[LabelChart]
		}
		out = append(out, plan.Observed{
			Name:                rel.Name,
			Namespace:           rel.Namespace,
			Chart:               chartName,
			Version:             rel.Chart.Metadata.Version,
			ValuesDigest:        rel.Labels[LabelValuesDigest],
			Status:              string(rel.Info.Status),
			Revision:            rel.Version,
			ManagedByDeployment: true,
		})
	}
	return out, nil
}

// InstallOrUpgrade applies one release.
func (c *Client) InstallOrUpgrade(ctx context.Context, rel ReleaseRequest, chartRef string, values map[string]any, digest string) error {
	_ = ctx
	exists, err := c.releaseExists(rel.Name, rel.Namespace)
	if err != nil {
		return err
	}
	labels := labels(c.deployment, rel.Chart, digest)
	if !exists {
		ins := action.NewInstall(c.cfg)
		ins.ReleaseName = rel.Name
		ins.Namespace = rel.Namespace
		ins.CreateNamespace = false
		ins.Timeout = c.timeout
		ins.WaitStrategy = waitStrategy(c.wait)
		ins.RollbackOnFailure = c.rollback
		ins.PostRenderer = pvcKeepRenderer()
		ins.Labels = labels
		ins.ChartPathOptions.Version = rel.Version
		ins.ChartPathOptions.RepoURL = chartRef
		ins.ChartPathOptions.PlainHTTP = c.regOpts.PlainHTTP
		if c.reg != nil {
			ins.SetRegistryClient(c.reg)
		}
		ch, err := c.locateChart(chartRef, rel.Version)
		if err != nil {
			return err
		}
		_, err = ins.RunWithContext(context.Background(), ch, values)
		if err == nil {
			return c.patchLabels(rel.Name, rel.Namespace, labels)
		}
		return err
	}
	up := action.NewUpgrade(c.cfg)
	up.Namespace = rel.Namespace
	up.Timeout = c.timeout
	up.WaitStrategy = waitStrategy(c.wait)
	up.RollbackOnFailure = c.rollback
	up.PostRenderer = pvcKeepRenderer()
	up.Labels = labels
	up.MaxHistory = c.historyMax
	up.ChartPathOptions.Version = rel.Version
	up.ChartPathOptions.RepoURL = chartRef
	up.ChartPathOptions.PlainHTTP = c.regOpts.PlainHTTP
	if c.reg != nil {
		up.SetRegistryClient(c.reg)
	}
	ch, err := c.locateChart(chartRef, rel.Version)
	if err != nil {
		return err
	}
	_, err = up.RunWithContext(context.Background(), rel.Name, ch, values)
	if err != nil {
		return err
	}
	return c.patchLabels(rel.Name, rel.Namespace, labels)
}

func (c *Client) locateChart(chartRef, version string) (any, error) {
	ins := action.NewInstall(c.cfg)
	ins.ChartPathOptions.Version = version
	ins.ChartPathOptions.RepoURL = chartRef
	ins.ChartPathOptions.PlainHTTP = c.regOpts.PlainHTTP
	if c.reg != nil {
		ins.SetRegistryClient(c.reg)
	}
	name, err := ins.ChartPathOptions.LocateChart(chartRef, c.settings)
	if err != nil {
		return nil, err
	}
	return loader.Load(name)
}

func (c *Client) releaseExists(name, ns string) (bool, error) {
	get := action.NewGet(c.cfg)
	get.Version = 0
	rel, err := get.Run(name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	r, ok := rel.(*release.Release)
	if !ok || r.Namespace != ns {
		return false, nil
	}
	return true, nil
}

func (c *Client) patchLabels(name, ns string, lbl map[string]string) error {
	_ = name
	_ = ns
	_ = lbl
	return nil
}

// Uninstall removes a release (PVC keep handled via post-render in C4).
func (c *Client) Uninstall(ctx context.Context, name, ns string) error {
	_ = ctx
	_ = ns
	un := action.NewUninstall(c.cfg)
	un.WaitStrategy = waitStrategy(true)
	un.Timeout = c.timeout
	_, err := un.Run(name)
	return err
}

// Rollback rolls back to previous revision.
func (c *Client) Rollback(ctx context.Context, name string, rev int) error {
	_ = ctx
	rb := action.NewRollback(c.cfg)
	rb.WaitStrategy = waitStrategy(true)
	rb.Timeout = c.timeout
	rb.Version = rev
	return rb.Run(name)
}

// ReleaseByName returns live release status.
func (c *Client) ReleaseByName(name string) (*release.Release, error) {
	get := action.NewGet(c.cfg)
	rel, err := get.Run(name)
	if err != nil {
		return nil, err
	}
	r, ok := rel.(*release.Release)
	if !ok {
		return nil, fmt.Errorf("unexpected release type for %q", name)
	}
	return r, nil
}

func waitStrategy(wait bool) helmkube.WaitStrategy {
	if wait {
		return helmkube.StatusWatcherStrategy
	}
	return helmkube.HookOnlyStrategy
}