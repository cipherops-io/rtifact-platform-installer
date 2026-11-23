package helmclient

import (
	"context"
	"os"
	"strings"
	"testing"

	chart "helm.sh/helm/v4/pkg/chart/v2"
	chartutil "helm.sh/helm/v4/pkg/chart/v2/util"
	"helm.sh/helm/v4/pkg/chart/common"
)

func TestPlannedManifest_leavesSharedConfigLive(t *testing.T) {
	hc := clientWithHistory(t)
	ch := &chart.Chart{
		Metadata:  &chart.Metadata{APIVersion: chart.APIVersionV2, Name: "app", Version: "1.0.0"},
		Templates: []*common.File{{Name: "templates/cm.yaml", Data: []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: app\n")}},
	}
	path, err := chartutil.Save(ch, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	hc.charts = NewChartCache()
	hc.charts.put("oci://r/app", "1.0.0", b)

	shared := hc.cfgByNS[testNS]
	releases, kubeClient := shared.Releases, shared.KubeClient
	manifest, err := hc.PlannedManifest(context.Background(), ReleaseRequest{Name: "app", Namespace: testNS}, "oci://r/app", "1.0.0", nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(manifest, "kind: ConfigMap") {
		t.Fatalf("manifest %q", manifest)
	}
	if shared.Releases != releases || shared.KubeClient != kubeClient {
		t.Fatal("dry-run replaced the shared configuration's storage or kube client; later installs would not reach the cluster")
	}
}
