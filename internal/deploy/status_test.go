package deploy

import (
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
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
	if ghost == nil || !ghost.InventoryOnly || ghost.Chart != "ghost" {
		t.Fatalf("ghost: %#v", ghost)
	}
}