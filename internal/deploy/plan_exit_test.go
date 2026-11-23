package deploy

import (
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
)

func TestPlan_exitContract_refusedBlocksDetailedExit(t *testing.T) {
	res := plan.Build(plan.Input{
		Wanted: nil,
		Inventory: []plan.Entry{
			{Name: "a", Namespace: "ns"},
			{Name: "b", Namespace: "ns"},
		},
	}, plan.Options{})
	if err := refusedBlocking(res); err == nil {
		t.Fatal("expected mass removal refused when file is empty")
	}
	if res.HasChange {
		t.Fatal("refused removals do not count as pending change for exit 2")
	}
}

func TestPlan_exitContract_downgradeRefused(t *testing.T) {
	res := plan.Build(plan.Input{
		Wanted: []plan.Wanted{{
			Name: "x", Namespace: "ns", Chart: "x", Version: "1.0.0", ValuesDigest: "d",
		}},
		Observed: []plan.Observed{{
			Name: "x", Namespace: "ns", Chart: "x", Version: "2.0.0", ValuesDigest: "d",
			Status: "deployed", ManagedByDeployment: true,
		}},
	}, plan.Options{})
	if err := refusedBlocking(res); err == nil {
		t.Fatal("expected downgrade refused")
	}
}

func TestPlan_exitContract_upgradePendingNoRefusal(t *testing.T) {
	res := plan.Build(plan.Input{
		Wanted: []plan.Wanted{{
			Name: "x", Namespace: "ns", Chart: "x", Version: "2.0.0", ValuesDigest: "d",
		}},
		Observed: []plan.Observed{{
			Name: "x", Namespace: "ns", Chart: "x", Version: "1.0.0", ValuesDigest: "d",
			Status: "deployed", ManagedByDeployment: true,
		}},
	}, plan.Options{})
	if err := refusedBlocking(res); err != nil {
		t.Fatal(err)
	}
	if !res.HasChange {
		t.Fatal("expected has change for upgrade")
	}
}