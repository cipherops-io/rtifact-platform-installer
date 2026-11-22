package plan

import "testing"

func TestBuildInstallAndUnchanged(t *testing.T) {
	in := Input{
		Wanted: []Wanted{
			{Name: "ok-a", Namespace: "ns", Chart: "ok-a", Version: "1.0.0", ValuesDigest: "aaa"},
		},
		Observed: []Observed{},
	}
	res := Build(in, Options{})
	if !res.HasChange || len(res.Rows) != 1 || res.Rows[0].Action != ActionInstall {
		t.Fatalf("expected install, got %#v", res)
	}

	in.Observed = []Observed{{
		Name: "ok-a", Namespace: "ns", Chart: "ok-a", Version: "1.0.0",
		ValuesDigest: "aaa", Status: "deployed", ManagedByDeployment: true,
	}}
	res = Build(in, Options{})
	if res.HasChange || res.Rows[0].Action != ActionUnchanged {
		t.Fatalf("expected unchanged, got %#v", res)
	}
}

func TestDowngradeRefused(t *testing.T) {
	in := Input{
		Wanted: []Wanted{{Name: "x", Namespace: "ns", Chart: "x", Version: "1.0.0", ValuesDigest: "d"}},
		Observed: []Observed{{
			Name: "x", Namespace: "ns", Chart: "x", Version: "2.0.0", ValuesDigest: "d",
			Status: "deployed", ManagedByDeployment: true,
		}},
	}
	res := Build(in, Options{})
	if res.Rows[0].Action != ActionRefused {
		t.Fatalf("expected refused downgrade, got %#v", res.Rows[0])
	}
}