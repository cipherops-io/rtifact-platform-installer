package deploy

import (
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
)

func TestMergeStatusRows_inventoryOnly(t *testing.T) {
	doc := &Document{Releases: []Release{
		{Name: "a", Namespace: "ns-a", Chart: "a", Version: "1.0.0"},
	}}
	inv := []inventory.Entry{
		{Name: "a", Namespace: "ns-a", Chart: "a", Version: "1.0.0"},
		{Name: "ghost", Namespace: "ns-g", Chart: "ghost", Version: "2.0.0"},
	}
	rows := mergeStatusRows(doc, inv, nil)
	if len(rows) != 2 {
		t.Fatalf("rows: %#v", rows)
	}
	var ghost *statusRow
	for i := range rows {
		if rows[i].Name == "ghost" {
			ghost = &rows[i]
		}
		if rows[i].Name == "a" && rows[i].InventoryOnly {
			t.Fatalf("file release must not be inventory-only: %#v", rows[i])
		}
	}
	if ghost == nil || !ghost.InventoryOnly || ghost.Chart != "ghost" || ghost.State != StateOrphaned {
		t.Fatalf("ghost: %#v", ghost)
	}
}

func TestMergeStatusRows_states(t *testing.T) {
	rel := func(name string) Release {
		return Release{Name: name, Namespace: "ns", Chart: name, Version: "1.0.0"}
	}
	doc := &Document{Releases: []Release{
		rel("ok"), rel("old"), rel("absent"), rel("busy"), rel("broke"), rel("foreign"),
	}}
	live := func(name, version, status string, managed bool) plan.Observed {
		return plan.Observed{Name: name, Namespace: "ns", Version: version, Status: status, ManagedByDeployment: managed}
	}
	rows := mergeStatusRows(doc, nil, []plan.Observed{
		live("ok", "1.0.0", "deployed", true),
		live("old", "0.9.0", "deployed", true),
		live("busy", "1.0.0", "pending-upgrade", true),
		live("broke", "1.0.0", "failed", true),
		live("foreign", "1.0.0", "deployed", false),
	})
	want := map[string]string{
		"ok": StateDeployed, "old": StateDrifted, "absent": StateMissing,
		"busy": StatePending, "broke": StateFailed, "foreign": StateUnmanaged,
	}
	for _, r := range rows {
		if r.State != want[r.Name] {
			t.Errorf("%s: state %q want %q", r.Name, r.State, want[r.Name])
		}
	}
}
