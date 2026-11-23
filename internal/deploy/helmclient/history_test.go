package helmclient

import (
	"context"
	"errors"
	"io"
	"testing"

	kubefake "helm.sh/helm/v4/pkg/kube/fake"
	releasecommon "helm.sh/helm/v4/pkg/release/common"
	release "helm.sh/helm/v4/pkg/release/v1"
	"helm.sh/helm/v4/pkg/storage"
	"helm.sh/helm/v4/pkg/storage/driver"

	"helm.sh/helm/v4/pkg/action"
)

const testNS = "ns"

func clientWithHistory(t *testing.T, statuses ...releasecommon.Status) *Client {
	t.Helper()
	cfg := &action.Configuration{
		Releases:   storage.Init(driver.NewMemory()),
		KubeClient: &kubefake.PrintingKubeClient{Out: io.Discard},
	}
	for i, st := range statuses {
		r := &release.Release{Name: "app", Namespace: testNS, Version: i + 1, Info: &release.Info{Status: st}}
		if err := cfg.Releases.Create(r); err != nil {
			t.Fatal(err)
		}
	}
	return &Client{cfgByNS: map[string]*action.Configuration{testNS: cfg}}
}

func TestPreviousDeployedRevision(t *testing.T) {
	cases := []struct {
		name    string
		history []releasecommon.Status
		want    int
	}{
		{"healthy upgrade rolls back to superseded", []releasecommon.Status{releasecommon.StatusSuperseded, releasecommon.StatusDeployed}, 1},
		{"failed upgrade rolls back to deployed", []releasecommon.Status{releasecommon.StatusDeployed, releasecommon.StatusFailed}, 1},
		{"interrupted upgrade", []releasecommon.Status{releasecommon.StatusSuperseded, releasecommon.StatusDeployed, releasecommon.StatusPendingUpgrade}, 2},
		{"skips failed revisions", []releasecommon.Status{releasecommon.StatusSuperseded, releasecommon.StatusFailed, releasecommon.StatusDeployed}, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := clientWithHistory(t, c.history...).PreviousDeployedRevision(context.Background(), "app", testNS)
			if err != nil || got != c.want {
				t.Fatalf("got %d, %v; want %d", got, err, c.want)
			}
		})
	}
}

func TestPreviousDeployedRevision_noneIsSentinel(t *testing.T) {
	for _, h := range [][]releasecommon.Status{
		{releasecommon.StatusDeployed},
		{releasecommon.StatusFailed},
		{releasecommon.StatusFailed, releasecommon.StatusFailed},
	} {
		_, err := clientWithHistory(t, h...).PreviousDeployedRevision(context.Background(), "app", testNS)
		if !errors.Is(err, ErrNoPreviousDeployedRevision) {
			t.Fatalf("history %v: want ErrNoPreviousDeployedRevision, got %v", h, err)
		}
	}
}

func TestFailedWithoutDeployed(t *testing.T) {
	cases := []struct {
		history []releasecommon.Status
		want    bool
	}{
		{[]releasecommon.Status{releasecommon.StatusFailed}, true},
		{[]releasecommon.Status{releasecommon.StatusFailed, releasecommon.StatusFailed}, true},
		{[]releasecommon.Status{releasecommon.StatusDeployed, releasecommon.StatusFailed}, false},
		{[]releasecommon.Status{releasecommon.StatusDeployed}, false},
		{nil, false},
	}
	for _, c := range cases {
		hc := clientWithHistory(t, c.history...)
		got, err := hc.failedWithoutDeployed(hc.cfgByNS[testNS], "app", testNS)
		if err != nil || got != c.want {
			t.Errorf("history %v: got %v, %v; want %v", c.history, got, err, c.want)
		}
	}
}

func TestLookup_reportsLatestRevision(t *testing.T) {
	hc := clientWithHistory(t, releasecommon.StatusSuperseded, releasecommon.StatusPendingInstall)
	info, err := hc.Lookup(context.Background(), "app", testNS)
	if err != nil {
		t.Fatal(err)
	}
	if info.Revision != 2 || !info.PendingInstall() {
		t.Fatalf("info %+v", info)
	}
	if _, err := clientWithHistory(t).Lookup(context.Background(), "app", testNS); !errors.Is(err, driver.ErrReleaseNotFound) {
		t.Fatalf("missing release: %v", err)
	}
	if manifest, err := clientWithHistory(t).ReleaseManifest(context.Background(), "app", testNS); err != nil || manifest != "" {
		t.Fatalf("missing release manifest: %q, %v", manifest, err)
	}
}

func TestChartCache(t *testing.T) {
	var nilCache *ChartCache
	nilCache.put("oci://r/c", "1.0.0", []byte("x"))
	if _, ok := nilCache.get("oci://r/c", "1.0.0"); ok {
		t.Fatal("nil cache must miss")
	}
	c := NewChartCache()
	c.put("oci://r/c", "1.0.0", []byte("v1"))
	if b, ok := c.get("oci://r/c", "1.0.0"); !ok || string(b) != "v1" {
		t.Fatalf("hit: %q %v", b, ok)
	}
	if _, ok := c.get("oci://r/c", "1.0.1"); ok {
		t.Fatal("different version must miss")
	}
	hc := &Client{charts: c}
	if b, err := hc.chartArchive("oci://r/c", "1.0.0"); err != nil || string(b) != "v1" {
		t.Fatalf("chartArchive should serve from cache without a registry: %q, %v", b, err)
	}
}
