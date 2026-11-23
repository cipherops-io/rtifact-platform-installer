package deploy

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/ocireg"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
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

func TestApply_refusalOnlyPlanFails(t *testing.T) {
	res := plan.Build(plan.Input{
		Inventory: []plan.Entry{{Name: "a", Namespace: "ns", Protect: true}},
	}, plan.Options{})
	var out bytes.Buffer
	ar := newApplyRun(RunOptions{Stdout: &out}, &Session{Doc: &Document{}}, &planRun{res: res})
	if err := ar.run(context.Background()); err == nil {
		t.Fatalf("apply of a plan with only refusals must fail; output:\n%s", out.String())
	}
	if strings.Contains(out.String(), "No changes.") {
		t.Fatalf("refusal reported as no changes:\n%s", out.String())
	}
}

func TestApply_noChangeStillRefreshesPullSecret(t *testing.T) {
	ctx := context.Background()
	cs := fake.NewSimpleClientset()
	doc := &Document{
		Registry:   Registry{Host: "reg.example", Auth: &RegistryAuth{UsernameEnv: "T_REG_USER", PasswordEnv: "T_REG_PASS"}},
		Namespaces: []Namespace{{Name: "ns", PullSecret: true}},
	}
	regCfg := ocireg.Registry{Host: "reg.example", UsernameEnv: "T_REG_USER", PasswordEnv: "T_REG_PASS"}
	apply := func() []byte {
		t.Helper()
		var out bytes.Buffer
		ar := newApplyRun(RunOptions{Stdout: &out}, &Session{Doc: doc, Client: cs}, &planRun{regCfg: regCfg})
		if err := ar.run(ctx); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), "No changes.") {
			t.Fatalf("output:\n%s", out.String())
		}
		sec, err := cs.CoreV1().Secrets("ns").Get(ctx, pullSecretName(doc), metav1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}
		return sec.Data[".dockerconfigjson"]
	}
	t.Setenv("T_REG_USER", "u")
	t.Setenv("T_REG_PASS", "old")
	before := apply()
	t.Setenv("T_REG_PASS", "rotated")
	if after := apply(); bytes.Equal(before, after) {
		t.Fatal("pull secret not refreshed after credential rotation")
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