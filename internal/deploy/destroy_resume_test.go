package deploy

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient/helmfake"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDestroyRun_resumesAfterMidListFailure(t *testing.T) {
	ctx := context.Background()
	cs := fake.NewSimpleClientset()
	const dep = "demo-dev"
	inv := []inventory.Entry{
		{Name: "base", Namespace: "ns", Chart: "base", Version: "1.0.0"},
		{Name: "mid", Namespace: "ns", Chart: "mid", Version: "1.0.0", Requires: []string{"base"}},
		{Name: "top", Namespace: "ns", Chart: "top", Version: "1.0.0", Requires: []string{"mid"}},
	}
	if err := inventory.Write(ctx, cs, dep, inv); err != nil {
		t.Fatal(err)
	}
	hc := helmfake.New(
		helmfake.Release{Name: "base", Namespace: "ns", Status: "deployed"},
		helmfake.Release{Name: "mid", Namespace: "ns", Status: "deployed"},
		helmfake.Release{Name: "top", Namespace: "ns", Status: "deployed"},
	)
	hc.FailUninstall["ns/mid"] = errors.New("boom")
	d := destroyRun{cs: cs, helm: hc, doc: &Document{}, deployment: dep, out: &bytes.Buffer{}}

	if err := d.run(ctx); err == nil {
		t.Fatal("expected destroy to fail on mid")
	}
	left, err := inventory.Read(ctx, cs, dep)
	if err != nil {
		t.Fatal(err)
	}
	if names := entryNames(left); len(names) != 2 || names["top"] || !names["mid"] || !names["base"] {
		t.Fatalf("after failure inventory should keep mid and base only: %v", names)
	}

	delete(hc.FailUninstall, "ns/mid")
	if err := d.run(ctx); err != nil {
		t.Fatalf("resume: %v", err)
	}
	left, err = inventory.Read(ctx, cs, dep)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) != 0 {
		t.Fatalf("inventory after resume: %#v", left)
	}
	for _, n := range []string{"base", "mid", "top"} {
		if hc.Has(n, "ns") {
			t.Fatalf("%s still installed", n)
		}
	}
}

func TestDestroyRun_protectedRefusedBeforeUninstall(t *testing.T) {
	ctx := context.Background()
	cs := fake.NewSimpleClientset()
	if err := inventory.Write(ctx, cs, "d", []inventory.Entry{{Name: "db", Namespace: "ns", Protect: true}}); err != nil {
		t.Fatal(err)
	}
	hc := helmfake.New(helmfake.Release{Name: "db", Namespace: "ns"})
	err := destroyRun{cs: cs, helm: hc, doc: &Document{}, deployment: "d", out: &bytes.Buffer{}}.run(ctx)
	if err == nil {
		t.Fatal("expected protected refusal")
	}
	if len(hc.Calls) != 0 {
		t.Fatalf("no helm calls expected, got %v", hc.Calls)
	}
}

func entryNames(es []inventory.Entry) map[string]bool {
	out := map[string]bool{}
	for _, e := range es {
		out[e.Name] = true
	}
	return out
}
