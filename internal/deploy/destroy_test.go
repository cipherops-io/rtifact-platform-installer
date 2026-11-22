package deploy

import (
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
)

func TestDestroyOrder_matchesPlanRemovalOrder(t *testing.T) {
	inv := []inventory.Entry{
		{Name: "db", Namespace: "ns", Version: "1.0.0"},
		{Name: "app", Namespace: "ns", Version: "1.0.0", Requires: []string{"db"}},
	}
	ordered := plan.RemovalOrder(inv)
	if len(ordered) != 2 || ordered[0].Name != "app" || ordered[1].Name != "db" {
		t.Fatalf("order: %#v", ordered)
	}
}