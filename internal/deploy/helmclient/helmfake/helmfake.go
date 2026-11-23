// Package helmfake is an in-memory helmclient.Interface for deploy unit tests.
package helmfake

import (
	"context"
	"fmt"
	"sync"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
)

// Release is one stored release.
type Release struct {
	Name, Namespace string
	Chart, Version  string
	Revision        int
	Status          string
	Labels          map[string]string
	Manifest        string
	ValuesDigest    string
}

// Client records calls and serves releases from memory.
type Client struct {
	mu       sync.Mutex
	releases map[string]*Release
	// FailUninstall makes Uninstall fail for the named release (namespace/name).
	FailUninstall map[string]error
	// PreviousRevision is the rollback target per release (namespace/name); absent means none.
	PreviousRevision map[string]int
	// Calls lists operations in order, e.g. "uninstall ns/name".
	Calls []string
}

var _ helmclient.Interface = (*Client)(nil)

// New returns a fake seeded with releases.
func New(rels ...Release) *Client {
	c := &Client{releases: map[string]*Release{}, FailUninstall: map[string]error{}, PreviousRevision: map[string]int{}}
	for i := range rels {
		r := rels[i]
		c.releases[key(r.Namespace, r.Name)] = &r
	}
	return c
}

func key(ns, name string) string { return ns + "/" + name }

// Has reports whether name exists in ns.
func (c *Client) Has(name, ns string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.releases[key(ns, name)]
	return ok
}

func (c *Client) record(format string, args ...any) {
	c.Calls = append(c.Calls, fmt.Sprintf(format, args...))
}

func (c *Client) Observe(_ context.Context, opts helmclient.ObserveOptions) ([]plan.Observed, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []plan.Observed
	for _, w := range opts.Wanted {
		r, ok := c.releases[key(w.Namespace, w.Name)]
		if !ok {
			continue
		}
		out = append(out, plan.Observed{
			Name: r.Name, Namespace: r.Namespace, Chart: r.Chart, Version: r.Version,
			ValuesDigest: r.ValuesDigest, Status: r.Status, Revision: r.Revision,
			ManagedByDeployment: r.Labels[helmclient.LabelDeployment] != "",
		})
	}
	return out, nil
}

func (c *Client) InstallOrUpgrade(_ context.Context, rel helmclient.ReleaseRequest, _ string, values map[string]any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.record("install %s/%s@%s", rel.Namespace, rel.Name, rel.Version)
	digest, err := plan.ValuesDigest(values)
	if err != nil {
		return err
	}
	cur, ok := c.releases[key(rel.Namespace, rel.Name)]
	rev := 1
	if ok {
		rev = cur.Revision + 1
	}
	c.releases[key(rel.Namespace, rel.Name)] = &Release{
		Name: rel.Name, Namespace: rel.Namespace, Chart: rel.Chart, Version: rel.Version,
		Revision: rev, Status: "deployed", ValuesDigest: digest,
		Labels: map[string]string{helmclient.LabelDeployment: "fake"},
	}
	return nil
}

func (c *Client) Uninstall(_ context.Context, name, ns string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.record("uninstall %s/%s", ns, name)
	if err := c.FailUninstall[key(ns, name)]; err != nil {
		return err
	}
	delete(c.releases, key(ns, name))
	return nil
}

func (c *Client) FetchChartArchive(context.Context, string, string) ([]byte, error) {
	return nil, fmt.Errorf("helmfake: FetchChartArchive not supported")
}

func (c *Client) PlannedManifest(context.Context, helmclient.ReleaseRequest, string, string, map[string]any, bool) (string, error) {
	return "", nil
}

func (c *Client) ReleaseManifest(_ context.Context, name, ns string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if r, ok := c.releases[key(ns, name)]; ok {
		return r.Manifest, nil
	}
	return "", nil
}

func (c *Client) Lookup(_ context.Context, name, ns string) (*helmclient.ReleaseInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	r, ok := c.releases[key(ns, name)]
	if !ok {
		return nil, fmt.Errorf("release: not found")
	}
	return &helmclient.ReleaseInfo{
		Name: r.Name, Namespace: r.Namespace, Revision: r.Revision, Status: r.Status, Labels: r.Labels,
	}, nil
}

func (c *Client) PreviousDeployedRevision(_ context.Context, name, ns string) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if rev, ok := c.PreviousRevision[key(ns, name)]; ok {
		return rev, nil
	}
	return 0, helmclient.ErrNoPreviousDeployedRevision
}

func (c *Client) Rollback(_ context.Context, name, ns string, rev int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.record("rollback %s/%s to %d", ns, name, rev)
	return nil
}
