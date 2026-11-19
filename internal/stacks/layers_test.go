package stacks

import (
	"reflect"
	"testing"
)

func TestApplyOrderHasFiveLayersAfterStateBackend(t *testing.T) {
	want := []Layer{"core-network", "data-platform", "app-platform", "network-edge", "ai-integrations"}
	if !reflect.DeepEqual(ApplyOrder, want) {
		t.Fatalf("ApplyOrder = %v, want %v", ApplyOrder, want)
	}
	all := AllLayers()
	if len(all) != 6 || all[0] != LayerStateBackend {
		t.Fatalf("AllLayers = %v", all)
	}
}

func TestParseAndResolve(t *testing.T) {
	for _, l := range AllLayers() {
		if got, ok := ParseLayer(string(l)); !ok || got != l {
			t.Fatalf("ParseLayer(%q) = %q, %v", l, got, ok)
		}
	}
	if _, ok := ParseLayer("network"); ok {
		t.Fatal("legacy layer name must not parse")
	}
	if got := ResolveLayers(LayerAll); len(got) != 5 {
		t.Fatalf("ResolveLayers(all) = %v", got)
	}
	if got := DestroyOrder(); got[0] != LayerAIIntegration || got[4] != LayerCoreNetwork {
		t.Fatalf("DestroyOrder = %v", got)
	}
}

func TestUpstream(t *testing.T) {
	if got := Upstream(LayerCoreNetwork); len(got) != 0 {
		t.Fatalf("core-network upstream = %v", got)
	}
	if got := Upstream(LayerDataPlatform); !reflect.DeepEqual(got, []Layer{LayerCoreNetwork}) {
		t.Fatalf("data-platform upstream = %v", got)
	}
	if got := Upstream(LayerStateBackend); len(got) != 0 {
		t.Fatalf("state-backend upstream = %v", got)
	}
}
