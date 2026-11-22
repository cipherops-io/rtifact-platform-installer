package deploy

import "testing"

func TestInstallOrder_requires(t *testing.T) {
	releases := []Release{
		{Name: "ok-b", Requires: []string{"ok-a"}},
		{Name: "ok-a"},
	}
	order, err := InstallOrder(releases)
	if err != nil {
		t.Fatal(err)
	}
	if order[0] != "ok-a" || order[1] != "ok-b" {
		t.Fatalf("order: %v", order)
	}
}

func TestInstallOrder_cycle(t *testing.T) {
	releases := []Release{
		{Name: "a", Requires: []string{"b"}},
		{Name: "b", Requires: []string{"a"}},
	}
	if _, err := InstallOrder(releases); err == nil {
		t.Fatal("expected cycle error")
	}
}