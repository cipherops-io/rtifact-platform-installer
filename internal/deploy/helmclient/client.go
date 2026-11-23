package helmclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
	"helm.sh/helm/v4/pkg/action"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/chart/v2/loader"
	"helm.sh/helm/v4/pkg/cli"
	helmkube "helm.sh/helm/v4/pkg/kube"
	"helm.sh/helm/v4/pkg/registry"
	releasecommon "helm.sh/helm/v4/pkg/release/common"
	release "helm.sh/helm/v4/pkg/release/v1"
	"helm.sh/helm/v4/pkg/storage/driver"
)

const helmDriver = "secret"

// ErrNoPreviousDeployedRevision is returned when rollback has no deployed history target.
var ErrNoPreviousDeployedRevision = errors.New("no previous deployed revision")

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

// ObserveOptions selects releases to read from the cluster.
type ObserveOptions struct {
	Wanted []plan.Wanted
}

// Client wraps Helm SDK install/upgrade/get for deploy.
type Client struct {
	kube       *kubeAccess
	chartEnv   *cli.EnvSettings
	cfgMu      sync.Mutex
	cfgByNS    map[string]*action.Configuration
	reg        *registry.Client
	deployment string
	regOpts    RegistryOpts
	timeout    time.Duration
	wait       bool
	rollback   bool
	historyMax int
	charts     *ChartCache
}

// Options for Helm client construction.
type Options struct {
	Kube           kube.Config
	RegistryClient *registry.Client
	Defaults       Defaults
	Registry       RegistryOpts
	Deployment     string
	// Charts is the session chart cache; nil disables caching.
	Charts *ChartCache
}

// New builds a Helm client from explicit kubeconfig (not ambient KUBECONFIG or HELM_KUBEAPISERVER).
func New(opts Options) (*Client, error) {
	cacheRoot := filepath.Join(os.TempDir(), "rtifact-installer-helm")
	if err := os.MkdirAll(cacheRoot, 0o755); err != nil {
		return nil, err
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
		kube:       newKubeAccess(opts.Kube),
		chartEnv:   chartLocateEnv(cacheRoot),
		cfgByNS:    map[string]*action.Configuration{},
		reg:        opts.RegistryClient,
		deployment: opts.Deployment,
		regOpts:    opts.Registry,
		timeout:    tout,
		wait:       wait,
		rollback:   rb,
		historyMax: opts.Defaults.HistoryMax,
		charts:     opts.Charts,
	}, nil
}

func (c *Client) configFor(ns string) (*action.Configuration, error) {
	if ns == "" {
		ns = "default"
	}
	c.cfgMu.Lock()
	defer c.cfgMu.Unlock()
	if cfg, ok := c.cfgByNS[ns]; ok {
		return cfg, nil
	}
	cfg := action.NewConfiguration()
	if err := cfg.Init(c.kube.RESTClientGetter(ns), ns, helmDriver); err != nil {
		return nil, err
	}
	if c.reg != nil {
		cfg.RegistryClient = c.reg
	}
	c.cfgByNS[ns] = cfg
	return cfg, nil
}

// Observe loads cluster state for each wanted release name/namespace (no cluster-wide list).
func (c *Client) Observe(ctx context.Context, opts ObserveOptions) ([]plan.Observed, error) {
	var out []plan.Observed
	for _, w := range opts.Wanted {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		cfg, err := c.configFor(w.Namespace)
		if err != nil {
			return nil, err
		}
		get := action.NewGet(cfg)
		rel, err := get.Run(w.Name)
		if err != nil {
			if errors.Is(err, driver.ErrReleaseNotFound) {
				continue
			}
			return nil, err
		}
		r, ok := rel.(*release.Release)
		if !ok {
			continue
		}
		if r.Namespace != w.Namespace {
			continue
		}
		managed := r.Labels != nil && r.Labels[LabelDeployment] == c.deployment
		releaseValues := r.Config
		if releaseValues == nil {
			releaseValues = map[string]any{}
		}
		valuesDigest, err := plan.ValuesDigest(releaseValues)
		if err != nil {
			return nil, fmt.Errorf("values digest for release %q: %w", w.Name, err)
		}
		out = append(out, plan.Observed{
			Name:                r.Name,
			Namespace:           r.Namespace,
			Chart:               r.Chart.Metadata.Name,
			Version:             r.Chart.Metadata.Version,
			ValuesDigest:        valuesDigest,
			Status:              string(r.Info.Status),
			Revision:            r.Version,
			ManagedByDeployment: managed,
		})
	}
	return out, nil
}

// InstallOrUpgrade applies one release.
func (c *Client) InstallOrUpgrade(ctx context.Context, rel ReleaseRequest, chartRef string, values map[string]any) error {
	cfg, err := c.configFor(rel.Namespace)
	if err != nil {
		return err
	}
	ch, err := c.locateChart(cfg, chartRef, rel.Version)
	if err != nil {
		return err
	}
	exists, err := c.releaseExists(cfg, rel.Name, rel.Namespace)
	if err != nil {
		return err
	}
	if exists {
		failedOnly, err := c.failedWithoutDeployed(cfg, rel.Name, rel.Namespace)
		if err != nil {
			return err
		}
		if failedOnly {
			if err := c.Uninstall(ctx, rel.Name, rel.Namespace); err != nil {
				return err
			}
			exists = false
		}
	}
	if !exists {
		return c.install(ctx, cfg, rel, chartRef, ch, values)
	}
	return c.upgrade(ctx, cfg, rel, chartRef, ch, values)
}

func (c *Client) install(ctx context.Context, cfg *action.Configuration, rel ReleaseRequest, chartRef string, ch *chart.Chart, values map[string]any) error {
	ins := action.NewInstall(cfg)
	ins.ReleaseName = rel.Name
	ins.Namespace = rel.Namespace
	ins.CreateNamespace = false
	ins.Timeout = c.timeout
	ins.WaitStrategy = waitStrategy(c.wait)
	ins.RollbackOnFailure = c.rollback
	ins.PostRenderer = pvcKeepRenderer()
	ins.Labels = labels(c.deployment)
	ins.ChartPathOptions.Version = rel.Version
	ins.ChartPathOptions.PlainHTTP = c.regOpts.PlainHTTP
	if c.reg != nil {
		ins.SetRegistryClient(c.reg)
	}
	_, err := ins.RunWithContext(ctx, ch, values)
	return err
}

func (c *Client) upgrade(ctx context.Context, cfg *action.Configuration, rel ReleaseRequest, chartRef string, ch *chart.Chart, values map[string]any) error {
	up := action.NewUpgrade(cfg)
	up.Namespace = rel.Namespace
	up.Timeout = c.timeout
	up.WaitStrategy = waitStrategy(c.wait)
	up.RollbackOnFailure = c.rollback
	up.PostRenderer = pvcKeepRenderer()
	up.Labels = labels(c.deployment)
	up.MaxHistory = c.historyMax
	up.ChartPathOptions.Version = rel.Version
	up.ChartPathOptions.PlainHTTP = c.regOpts.PlainHTTP
	if c.reg != nil {
		up.SetRegistryClient(c.reg)
	}
	_, err := up.RunWithContext(ctx, rel.Name, ch, values)
	return err
}

func (c *Client) locateChart(_ *action.Configuration, chartRef, version string) (*chart.Chart, error) {
	b, err := c.chartArchive(chartRef, version)
	if err != nil {
		return nil, err
	}
	return loader.LoadArchive(bytes.NewReader(b))
}

// chartArchive downloads chartRef@version once per session and returns the .tgz bytes.
func (c *Client) chartArchive(chartRef, version string) ([]byte, error) {
	if b, ok := c.charts.get(chartRef, version); ok {
		return b, nil
	}
	cfg, err := c.configFor("default")
	if err != nil {
		return nil, err
	}
	ins := action.NewInstall(cfg)
	ins.ChartPathOptions.Version = version
	ins.ChartPathOptions.PlainHTTP = c.regOpts.PlainHTTP
	if c.reg != nil {
		ins.SetRegistryClient(c.reg)
	}
	path, err := ins.ChartPathOptions.LocateChart(chartRef, c.chartEnv)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c.charts.put(chartRef, version, b)
	return b, nil
}

func (c *Client) releaseExists(cfg *action.Configuration, name, ns string) (bool, error) {
	get := action.NewGet(cfg)
	get.Version = 0
	rel, err := get.Run(name)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
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

// Uninstall removes a release; PVCs survive via the keep post-renderer.
func (c *Client) Uninstall(ctx context.Context, name, ns string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	cfg, err := c.configFor(ns)
	if err != nil {
		return err
	}
	un := action.NewUninstall(cfg)
	un.WaitStrategy = waitStrategy(true)
	un.Timeout = c.timeout
	_, err = un.Run(name)
	if err != nil && errors.Is(err, driver.ErrReleaseNotFound) {
		return nil
	}
	return err
}

// FetchChartArchive returns the chart .tgz bytes (for CRD upgrades), using the session cache.
func (c *Client) FetchChartArchive(ctx context.Context, chartRef, version string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return c.chartArchive(chartRef, version)
}

// PreviousDeployedRevision returns the newest revision before the current one that
// was successfully deployed. Helm marks a deployed revision superseded once a later
// upgrade succeeds, so both statuses count (as in Helm's own rollback-on-failure).
func (c *Client) PreviousDeployedRevision(ctx context.Context, name, ns string) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	cfg, err := c.configFor(ns)
	if err != nil {
		return 0, err
	}
	get := action.NewGet(cfg)
	curRel, err := get.Run(name)
	if err != nil {
		return 0, err
	}
	cur, ok := curRel.(*release.Release)
	if !ok {
		return 0, fmt.Errorf("unexpected release type for %q", name)
	}
	if cur.Namespace != ns {
		return 0, fmt.Errorf("release %q is not in namespace %q", name, ns)
	}
	hist := action.NewHistory(cfg)
	revs, err := hist.Run(name)
	if err != nil {
		return 0, err
	}
	target := 0
	for i := len(revs) - 1; i >= 0; i-- {
		r, ok := revs[i].(*release.Release)
		if !ok {
			continue
		}
		if r.Version >= cur.Version {
			continue
		}
		if r.Info.Status == releasecommon.StatusDeployed || r.Info.Status == releasecommon.StatusSuperseded {
			target = r.Version
			break
		}
	}
	if target < 1 {
		return 0, fmt.Errorf("release %q: %w", name, ErrNoPreviousDeployedRevision)
	}
	return target, nil
}

func (c *Client) failedWithoutDeployed(cfg *action.Configuration, name, ns string) (bool, error) {
	get := action.NewGet(cfg)
	rel, err := get.Run(name)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return false, nil
		}
		return false, err
	}
	r, ok := rel.(*release.Release)
	if !ok || r.Namespace != ns {
		return false, nil
	}
	if r.Info.Status != releasecommon.StatusFailed {
		return false, nil
	}
	hist := action.NewHistory(cfg)
	revs, err := hist.Run(name)
	if err != nil {
		return false, err
	}
	for _, item := range revs {
		rev, ok := item.(*release.Release)
		if !ok {
			continue
		}
		if rev.Info.Status == releasecommon.StatusDeployed {
			return false, nil
		}
	}
	return true, nil
}

// Rollback rolls back to previous revision.
func (c *Client) Rollback(ctx context.Context, name, ns string, rev int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	cfg, err := c.configFor(ns)
	if err != nil {
		return err
	}
	rb := action.NewRollback(cfg)
	rb.WaitStrategy = waitStrategy(true)
	rb.Timeout = c.timeout
	rb.Version = rev
	return rb.Run(name)
}

// Lookup returns the latest revision of a release in ns.
func (c *Client) Lookup(ctx context.Context, name, ns string) (*ReleaseInfo, error) {
	r, err := c.release(ctx, name, ns)
	if err != nil {
		return nil, err
	}
	return releaseInfo(r), nil
}

func (c *Client) release(ctx context.Context, name, ns string) (*release.Release, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cfg, err := c.configFor(ns)
	if err != nil {
		return nil, err
	}
	get := action.NewGet(cfg)
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
