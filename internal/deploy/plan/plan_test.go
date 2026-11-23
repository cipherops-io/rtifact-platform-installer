package plan

import (
	"strings"
	"testing"
)

func w(name, version, vd string) Wanted {
	return Wanted{Name: name, Namespace: "ns", Chart: name, Version: version, ValuesDigest: vd}
}

func obs(name, version, vd, status string) Observed {
	return Observed{
		Name: name, Namespace: "ns", Chart: name, Version: version,
		ValuesDigest: vd, Status: status, ManagedByDeployment: true,
	}
}

func inv(name, version string, requires ...string) Entry {
	return Entry{Name: name, Namespace: "ns", Chart: name, Version: version, Requires: requires}
}

func rowFor(t *testing.T, res Result, name string) Row {
	t.Helper()
	for _, r := range res.Rows {
		if r.Name == name {
			return r
		}
	}
	t.Fatalf("no row for %q in %#v", name, res.Rows)
	return Row{}
}

func TestBuildInstallAndUnchanged(t *testing.T) {
	in := Input{Wanted: []Wanted{w("ok-a", "1.0.0", "aaa")}}
	res := Build(in, Options{})
	if !res.HasChange || len(res.Rows) != 1 || res.Rows[0].Action != ActionInstall {
		t.Fatalf("expected install, got %#v", res)
	}

	in.Observed = []Observed{obs("ok-a", "1.0.0", "aaa", "deployed")}
	res = Build(in, Options{})
	if res.HasChange || res.Rows[0].Action != ActionUnchanged {
		t.Fatalf("expected unchanged, got %#v", res)
	}
}

func TestUnchangedRequiresDeployedAndDigests(t *testing.T) {
	tests := []struct {
		name    string
		want    Wanted
		o       Observed
		inv     []Entry
		want_   Action
		noteHas string
	}{
		{"failed same version is not unchanged", w("x", "1.0.0", "d"), obs("x", "1.0.0", "d", "failed"), nil, ActionUpgrade, "failed"},
		{"superseded same version is not unchanged", w("x", "1.0.0", "d"), obs("x", "1.0.0", "d", "superseded"), nil, ActionUpgrade, "superseded"},
		{"values digest differs", w("x", "1.0.0", "new"), obs("x", "1.0.0", "old", "deployed"), nil, ActionUpgrade, "values"},
		{"values digest from observed matches", w("x", "1.0.0", "d"), obs("x", "1.0.0", "d", "deployed"), nil, ActionUnchanged, ""},
		{"inventory values digest ignored", w("x", "1.0.0", "d"), obs("x", "1.0.0", "d", "deployed"),
			[]Entry{{Name: "x", Namespace: "ns", Chart: "x", Version: "1.0.0", ValuesDigest: "old-file"}}, ActionUnchanged, ""},
		{"values digest unknown", w("x", "1.0.0", "d"), obs("x", "1.0.0", "", "deployed"), nil, ActionUpgrade, "values"},
		{"missing observed values digest", w("x", "1.0.0", "d"), obs("x", "1.0.0", "", "deployed"),
			[]Entry{{Name: "x", Namespace: "ns", Chart: "x", Version: "0.9.0", ValuesDigest: "d"}}, ActionUpgrade, "values"},
		{"chart digest differs same version", Wanted{Name: "x", Namespace: "ns", Chart: "x", Version: "1.0.0", ChartDigest: "sha256:new", ValuesDigest: "d"},
			obs("x", "1.0.0", "d", "deployed"),
			[]Entry{{Name: "x", Namespace: "ns", Chart: "x", Version: "1.0.0", ChartDigest: "sha256:old", ValuesDigest: "d"}}, ActionUpgrade, "chart digest"},
		{"chart digest matches", Wanted{Name: "x", Namespace: "ns", Chart: "x", Version: "1.0.0", ChartDigest: "sha256:a", ValuesDigest: "d"},
			obs("x", "1.0.0", "d", "deployed"),
			[]Entry{{Name: "x", Namespace: "ns", Chart: "x", Version: "1.0.0", ChartDigest: "sha256:a", ValuesDigest: "d"}}, ActionUnchanged, ""},
		{"chart digest unresolved is skipped", w("x", "1.0.0", "d"),
			func() Observed { o := obs("x", "1.0.0", "d", "deployed"); o.ChartDigest = "sha256:a"; return o }(), nil, ActionUnchanged, ""},
		{"chart name changed", Wanted{Name: "x", Namespace: "ns", Chart: "y", Version: "1.0.0", ValuesDigest: "d"}, obs("x", "1.0.0", "d", "deployed"), nil, ActionUpgrade, "chart changed"},
		{"semver equal spelling", w("x", "v1.0.0", "d"), obs("x", "1.0.0", "d", "deployed"), nil, ActionUnchanged, ""},
		{"upgrade 1.9.0 to 1.10.0", w("x", "1.10.0", "d"), obs("x", "1.9.0", "d", "deployed"), nil, ActionUpgrade, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := Build(Input{Wanted: []Wanted{tc.want}, Observed: []Observed{tc.o}, Inventory: tc.inv}, Options{})
			r := res.Rows[0]
			if r.Action != tc.want_ {
				t.Fatalf("action = %s (%s), want %s", r.Action, r.Note, tc.want_)
			}
			if !strings.Contains(r.Note, tc.noteHas) {
				t.Fatalf("note %q missing %q", r.Note, tc.noteHas)
			}
			if wantChange := tc.want_ != ActionUnchanged; res.HasChange != wantChange {
				t.Fatalf("HasChange = %v, want %v", res.HasChange, wantChange)
			}
		})
	}
}

func TestDowngradeSemver(t *testing.T) {
	// String comparison would call 1.9.0 > 1.10.0; semver says it is a downgrade.
	in := Input{
		Wanted:   []Wanted{w("x", "1.9.0", "d")},
		Observed: []Observed{obs("x", "1.10.0", "d", "deployed")},
	}
	res := Build(in, Options{})
	r := res.Rows[0]
	if r.Action != ActionRefused || !strings.Contains(r.Note, "--allow-downgrade") {
		t.Fatalf("expected refused downgrade, got %#v", r)
	}
	if res.HasChange {
		t.Fatalf("HasChange must be false when all rows refused")
	}

	res = Build(in, Options{AllowDowngrade: true})
	if res.Rows[0].Action != ActionUpgrade || !res.HasChange {
		t.Fatalf("expected downgrade allowed, got %#v", res.Rows[0])
	}
}

func TestInvalidSemverRefused(t *testing.T) {
	res := Build(Input{Wanted: []Wanted{w("x", "latest", "d")}}, Options{})
	if res.Rows[0].Action != ActionRefused || !strings.Contains(res.Rows[0].Note, "semver") {
		t.Fatalf("expected semver refusal, got %#v", res.Rows[0])
	}
	res = Build(Input{
		Wanted:   []Wanted{w("x", "1.0.0", "d")},
		Observed: []Observed{obs("x", "weird", "d", "deployed")},
	}, Options{})
	if res.Rows[0].Action != ActionRefused || !strings.Contains(res.Rows[0].Note, "semver") {
		t.Fatalf("expected semver refusal for observed, got %#v", res.Rows[0])
	}
}

func TestAdopt(t *testing.T) {
	unmanaged := obs("x", "1.0.0", "d", "deployed")
	unmanaged.ManagedByDeployment = false
	in := Input{Wanted: []Wanted{w("x", "1.0.0", "d")}, Observed: []Observed{unmanaged}}

	res := Build(in, Options{})
	if r := res.Rows[0]; r.Action != ActionRefused || !strings.Contains(r.Note, "--adopt") {
		t.Fatalf("expected refused without --adopt, got %#v", r)
	}
	if res.HasChange {
		t.Fatalf("refused adopt must not count as change")
	}

	res = Build(in, Options{AllowAdopt: true})
	if res.Rows[0].Action != ActionAdopt || !res.HasChange {
		t.Fatalf("expected adopt, got %#v", res.Rows[0])
	}
}

func TestStuckBlocksDependents(t *testing.T) {
	a, b, c, d := w("a", "1.0.0", "d"), w("b", "2.0.0", "d"), w("c", "1.0.0", "d"), w("d", "1.0.0", "d")
	b.Requires = []string{"a"}
	c.Requires = []string{"b"} // transitive
	d.Requires = nil
	in := Input{
		FilePath: "/deploy/releases.yaml",
		Wanted:   []Wanted{a, b, c, d},
		Observed: []Observed{
			obs("a", "1.0.0", "d", "pending-upgrade"),
			obs("b", "1.0.0", "d", "deployed"),
		},
	}
	res := Build(in, Options{})
	st := rowFor(t, res, "a")
	if st.Action != ActionStuck {
		t.Fatalf("a: %#v", st)
	}
	wantCmd := "rtifact deploy rollback /deploy/releases.yaml a"
	if !strings.Contains(st.Note, wantCmd) {
		t.Fatalf("note %q missing full rollback command %q", st.Note, wantCmd)
	}
	for _, name := range []string{"b", "c"} {
		r := rowFor(t, res, name)
		if r.Action != ActionRefused || !strings.Contains(r.Note, "blocked") {
			t.Fatalf("%s should be blocked by stuck a: %#v", name, r)
		}
	}
	if r := rowFor(t, res, "d"); r.Action != ActionInstall {
		t.Fatalf("independent release should proceed: %#v", r)
	}
}

func TestStuckNotePlaceholderWhenNoFile(t *testing.T) {
	res := Build(Input{
		Wanted:   []Wanted{w("x", "1.0.0", "d")},
		Observed: []Observed{obs("x", "1.0.0", "d", "pending-install")},
	}, Options{})
	if got := res.Rows[0].Note; !strings.Contains(got, "rtifact deploy rollback <deploy-file> x") {
		t.Fatalf("note %q", got)
	}
}

func TestHasChangeFalseWhenAllRefused(t *testing.T) {
	in := Input{
		Wanted:   []Wanted{w("x", "1.0.0", "d")},
		Observed: []Observed{obs("x", "2.0.0", "d", "deployed")},
		Inventory: []Entry{
			inv("x", "2.0.0"), {Name: "gone", Namespace: "ns", Chart: "gone", Version: "1.0.0", Protect: true},
			inv("keep", "1.0.0"),
		},
	}
	res := Build(in, Options{})
	for _, r := range res.Rows {
		if r.Action != ActionRefused {
			t.Fatalf("expected only refused rows, got %#v", res.Rows)
		}
	}
	if res.HasChange {
		t.Fatalf("HasChange must be false")
	}
}

// Removal guards (design: "Removal: six guards").

func TestGuard1_OnlyWhatInventoryListed(t *testing.T) {
	// "stranger" is live and labelled but not in inventory: never a candidate.
	// "orphan" is in inventory and absent from the file: candidate.
	in := Input{
		Wanted: []Wanted{w("keep", "1.0.0", "d"), w("k2", "1.0.0", "d"), w("k3", "1.0.0", "d")},
		Observed: []Observed{
			obs("keep", "1.0.0", "d", "deployed"),
			obs("stranger", "1.0.0", "d", "deployed"),
			obs("orphan", "1.0.0", "d", "deployed"),
		},
		Inventory: []Entry{inv("keep", "1.0.0"), inv("k2", "1.0.0"), inv("k3", "1.0.0"), inv("orphan", "1.0.0"), inv("other", "1.0.0")},
	}
	res := Build(in, Options{})
	if len(res.Removals) != 2 {
		t.Fatalf("removals = %#v", res.Removals)
	}
	for _, r := range res.Removals {
		if r.Name == "stranger" {
			t.Fatalf("release outside inventory must not be removed")
		}
	}
	if r := rowFor(t, res, "orphan"); r.Action != ActionRemove {
		t.Fatalf("orphan: %#v", r)
	}
}

func TestGuard1_UnmanagedLiveReleaseNotRemoved(t *testing.T) {
	foreign := obs("orphan", "1.0.0", "d", "deployed")
	foreign.ManagedByDeployment = false
	res := Build(Input{
		Wanted:    []Wanted{w("keep", "1.0.0", "d")},
		Observed:  []Observed{obs("keep", "1.0.0", "d", "deployed"), foreign},
		Inventory: []Entry{inv("keep", "1.0.0"), inv("orphan", "1.0.0"), inv("x", "1.0.0")},
	}, Options{})
	if r := rowFor(t, res, "orphan"); r.Action != ActionRefused || !strings.Contains(r.Note, "not managed") {
		t.Fatalf("orphan: %#v", r)
	}
}

func TestGuard2_RemovalsListedSeparatelyAfterRows(t *testing.T) {
	res := Build(Input{
		Wanted:    []Wanted{w("keep", "1.0.0", "d")},
		Observed:  []Observed{obs("keep", "1.0.0", "d", "deployed"), obs("old", "1.0.0", "d", "deployed")},
		Inventory: []Entry{inv("keep", "1.0.0"), inv("old", "1.0.0")},
	}, Options{})
	if len(res.Removals) != 1 || res.Removals[0].Name != "old" {
		t.Fatalf("removals = %#v", res.Removals)
	}
	if last := res.Rows[len(res.Rows)-1]; last.Name != "old" || last.Action != ActionRemove {
		t.Fatalf("removal must be listed last: %#v", res.Rows)
	}
	if !res.HasChange {
		t.Fatalf("removal is a change")
	}
}

func TestGuard3_DependentsBlockRemoval(t *testing.T) {
	app := w("app", "1.0.0", "d")
	app.Requires = []string{"db"}
	res := Build(Input{
		Wanted:    []Wanted{app},
		Observed:  []Observed{obs("app", "1.0.0", "d", "deployed"), obs("db", "1.0.0", "d", "deployed")},
		Inventory: []Entry{inv("app", "1.0.0", "db"), inv("db", "1.0.0"), inv("x", "1.0.0"), inv("y", "1.0.0")},
	}, Options{})
	r := rowFor(t, res, "db")
	if r.Action != ActionRefused || !strings.Contains(r.Note, "required by app") {
		t.Fatalf("db: %#v", r)
	}
}

func TestGuard4_ProtectNeedsAllowDestroy(t *testing.T) {
	e := inv("vault", "1.0.0")
	e.Protect = true
	in := Input{
		Wanted:    []Wanted{w("keep", "1.0.0", "d")},
		Observed:  []Observed{obs("keep", "1.0.0", "d", "deployed"), obs("vault", "1.0.0", "d", "deployed")},
		Inventory: []Entry{inv("keep", "1.0.0"), e},
	}
	res := Build(in, Options{})
	if r := rowFor(t, res, "vault"); r.Action != ActionRefused || !strings.Contains(r.Note, "--allow-destroy") {
		t.Fatalf("vault: %#v", r)
	}
	res = Build(in, Options{AllowDestroy: true})
	if r := rowFor(t, res, "vault"); r.Action != ActionRemove {
		t.Fatalf("vault with --allow-destroy: %#v", r)
	}
}

func TestGuard5_MassRemoval(t *testing.T) {
	inventory := []Entry{inv("a", "1.0.0"), inv("b", "1.0.0"), inv("c", "1.0.0"), inv("d", "1.0.0")}
	tests := []struct {
		name    string
		wanted  []Wanted
		opts    Options
		refused int
		removed int
	}{
		{"empty file blocked", nil, Options{}, 4, 0},
		{"empty file allowed", nil, Options{AllowMassRemoval: true}, 0, 4},
		{"3 of 4 blocked", []Wanted{w("a", "1.0.0", "d")}, Options{}, 3, 0},
		{"exactly half allowed", []Wanted{w("a", "1.0.0", "d"), w("b", "1.0.0", "d")}, Options{}, 0, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := Build(Input{Wanted: tc.wanted, Inventory: inventory}, tc.opts)
			var refused, removed int
			for _, r := range res.Removals {
				switch r.Action {
				case ActionRefused:
					refused++
					if !strings.Contains(r.Note, "--allow-mass-removal") {
						t.Fatalf("note %q", r.Note)
					}
				case ActionRemove:
					removed++
				}
			}
			if refused != tc.refused || removed != tc.removed {
				t.Fatalf("refused=%d removed=%d, want %d/%d: %#v", refused, removed, tc.refused, tc.removed, res.Removals)
			}
			if tc.removed == 0 && tc.wanted == nil && res.HasChange {
				t.Fatalf("HasChange must be false when everything is refused")
			}
		})
	}
}

func TestGuard5_InventoryOnlyOrphan(t *testing.T) {
	// Entry in inventory, release already gone from the cluster.
	res := Build(Input{
		Wanted:    []Wanted{w("a", "1.0.0", "d"), w("b", "1.0.0", "d")},
		Observed:  []Observed{obs("a", "1.0.0", "d", "deployed"), obs("b", "1.0.0", "d", "deployed")},
		Inventory: []Entry{inv("a", "1.0.0"), inv("b", "1.0.0"), inv("ghost", "1.0.0")},
	}, Options{})
	r := rowFor(t, res, "ghost")
	if r.Action != ActionRemove || !strings.Contains(r.Note, "inventory-only") {
		t.Fatalf("ghost: %#v", r)
	}
}

func TestGuard6_DataStaysUnlessPurge(t *testing.T) {
	in := Input{
		Wanted:    []Wanted{w("a", "1.0.0", "d"), w("b", "1.0.0", "d")},
		Observed:  []Observed{obs("a", "1.0.0", "d", "deployed"), obs("b", "1.0.0", "d", "deployed"), obs("old", "1.0.0", "d", "deployed")},
		Inventory: []Entry{inv("a", "1.0.0"), inv("b", "1.0.0"), inv("old", "1.0.0")},
	}
	res := Build(in, Options{})
	if n := rowFor(t, res, "old").Note; !strings.Contains(n, "kept") {
		t.Fatalf("note %q", n)
	}
	res = Build(in, Options{Purge: true})
	if n := rowFor(t, res, "old").Note; !strings.Contains(n, "--purge") || !strings.Contains(n, "deleted") {
		t.Fatalf("note %q", n)
	}
}

func TestRemovalOrderDependentsFirst(t *testing.T) {
	res := Build(Input{
		Wanted: []Wanted{w("keep", "1.0.0", "d")},
		Inventory: []Entry{
			inv("keep", "1.0.0"),
			inv("db", "1.0.0"),
			inv("app", "1.0.0", "db"),
		},
	}, Options{AllowMassRemoval: true})
	if len(res.Removals) != 2 || res.Removals[0].Name != "app" || res.Removals[1].Name != "db" {
		t.Fatalf("removals = %#v", res.Removals)
	}
}

func TestOnlyRestrictsRemovals(t *testing.T) {
	res := Build(Input{
		Wanted:    []Wanted{w("keep", "1.0.0", "d")},
		Inventory: []Entry{inv("keep", "1.0.0"), inv("a", "1.0.0"), inv("b", "1.0.0")},
	}, Options{Only: "a", OnlyClosure: map[string]bool{"a": true}})
	if len(res.Removals) != 1 || res.Removals[0].Name != "a" || res.Removals[0].Action != ActionRemove {
		t.Fatalf("--only a must consider only a: %#v", res.Removals)
	}
}

func TestRollbackCommandQuotesPath(t *testing.T) {
	got := RollbackCommand("/my files/r.yaml", "x")
	if got != "rtifact deploy rollback '/my files/r.yaml' x" {
		t.Fatalf("got %q", got)
	}
}
