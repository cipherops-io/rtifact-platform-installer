package helmclient

import (
	"context"
	"sync"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
	releasecommon "helm.sh/helm/v4/pkg/release/common"
	release "helm.sh/helm/v4/pkg/release/v1"
)

// Interface is the Helm surface the deploy engine uses; *Client implements it
// and helmfake provides an in-memory double for unit tests.
type Interface interface {
	Observe(ctx context.Context, opts ObserveOptions) ([]plan.Observed, error)
	InstallOrUpgrade(ctx context.Context, rel ReleaseRequest, chartRef string, values map[string]any) error
	Uninstall(ctx context.Context, name, ns string) error
	FetchChartArchive(ctx context.Context, chartRef, version string) ([]byte, error)
	PlannedManifest(ctx context.Context, rel ReleaseRequest, chartRef, version string, values map[string]any, upgrade bool) (string, error)
	ReleaseManifest(ctx context.Context, name, ns string) (string, error)
	Lookup(ctx context.Context, name, ns string) (*ReleaseInfo, error)
	PreviousDeployedRevision(ctx context.Context, name, ns string) (int, error)
	Rollback(ctx context.Context, name, ns string, rev int) error
}

var _ Interface = (*Client)(nil)

// ReleaseInfo is the latest revision of a release without Helm SDK types.
type ReleaseInfo struct {
	Name, Namespace string
	Revision        int
	Status          string
	Labels          map[string]string
}

// PendingInstall reports a first install that never completed (no revision to roll back to).
func (r *ReleaseInfo) PendingInstall() bool {
	return r != nil && r.Status == releasecommon.StatusPendingInstall.String()
}

func releaseInfo(r *release.Release) *ReleaseInfo {
	info := &ReleaseInfo{Name: r.Name, Namespace: r.Namespace, Revision: r.Version, Labels: r.Labels}
	if r.Info != nil {
		info.Status = r.Info.Status.String()
	}
	return info
}

// ChartCache keeps downloaded chart archives for one deploy session.
type ChartCache struct {
	mu   sync.Mutex
	byID map[string][]byte
}

// NewChartCache returns an empty per-session chart cache.
func NewChartCache() *ChartCache {
	return &ChartCache{byID: map[string][]byte{}}
}

func (c *ChartCache) get(ref, version string) ([]byte, bool) {
	if c == nil {
		return nil, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	b, ok := c.byID[ref+"@"+version]
	return b, ok
}

func (c *ChartCache) put(ref, version string, b []byte) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.byID[ref+"@"+version] = b
}
