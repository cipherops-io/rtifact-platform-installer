package helmclient

import (
	"context"
	"errors"
	"fmt"

	"helm.sh/helm/v4/pkg/action"
	release "helm.sh/helm/v4/pkg/release/v1"
	"helm.sh/helm/v4/pkg/storage/driver"
)

// PlannedManifest renders install/upgrade manifests without applying (Secrets masked when HideSecret is set).
func (c *Client) PlannedManifest(ctx context.Context, rel ReleaseRequest, chartRef, version string, values map[string]any, upgrade bool) (string, error) {
	cfg, err := c.configFor(rel.Namespace)
	if err != nil {
		return "", err
	}
	ch, err := c.locateChart(cfg, chartRef, version)
	if err != nil {
		return "", err
	}
	if upgrade {
		up := action.NewUpgrade(cfg)
		up.DryRunStrategy = action.DryRunClient
		up.HideSecret = true
		up.Namespace = rel.Namespace
		up.ChartPathOptions.Version = version
		up.ChartPathOptions.PlainHTTP = c.regOpts.PlainHTTP
		if c.reg != nil {
			up.SetRegistryClient(c.reg)
		}
		res, err := up.RunWithContext(ctx, rel.Name, ch, values)
		if err != nil {
			return "", err
		}
		return releaserManifest(res)
	}
	ins := action.NewInstall(cfg)
	ins.DryRunStrategy = action.DryRunClient
	ins.HideSecret = true
	ins.ReleaseName = rel.Name
	ins.Namespace = rel.Namespace
	ins.ChartPathOptions.Version = version
	ins.ChartPathOptions.PlainHTTP = c.regOpts.PlainHTTP
	if c.reg != nil {
		ins.SetRegistryClient(c.reg)
	}
	res, err := ins.RunWithContext(ctx, ch, values)
	if err != nil {
		return "", err
	}
	return releaserManifest(res)
}

// ReleaseManifest returns the stored manifest for a live release ("" when it does not exist).
func (c *Client) ReleaseManifest(ctx context.Context, name, ns string) (string, error) {
	rel, err := c.release(ctx, name, ns)
	if errors.Is(err, driver.ErrReleaseNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return rel.Manifest, nil
}

func releaserManifest(res any) (string, error) {
	r, ok := res.(*release.Release)
	if !ok {
		return "", fmt.Errorf("unexpected release type %T", res)
	}
	return r.Manifest, nil
}
