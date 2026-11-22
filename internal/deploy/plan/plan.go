package plan

// Action is the planned change for one release.
type Action string

const (
	ActionInstall    Action = "install"
	ActionUpgrade    Action = "upgrade"
	ActionUnchanged  Action = "unchanged"
	ActionRemove     Action = "remove"
	ActionAdopt      Action = "adopt"
	ActionStuck      Action = "stuck"
	ActionRefused    Action = "refused"
)

// Wanted describes desired state from the deploy file.
type Wanted struct {
	Name, Namespace string
	Chart, Version  string
	ValuesDigest    string
	Protect         bool
}

// Observed is live cluster state for a release managed by this deployment.
type Observed struct {
	Name, Namespace     string
	Chart, Version      string
	ValuesDigest        string
	Status              string // deployed, pending-upgrade, pending-install, failed, etc.
	Revision            int
	ManagedByDeployment bool
}

// Row is one line in a deploy plan.
type Row struct {
	Name, Namespace string
	Action          Action
	FromVersion     string
	ToVersion       string
	Note            string
}

// Input is pure data for planning (no cluster / Helm calls).
type Input struct {
	Wanted     []Wanted
	Observed   []Observed
	Inventory  []Wanted
	Deployment string
}

// Options toggles guard overrides.
type Options struct {
	AllowDestroy      bool
	AllowMassRemoval  bool
	AllowDowngrade    bool
	AllowAdopt        bool
	Only              string
	OnlyClosure       map[string]bool
}

// Result is the full plan.
type Result struct {
	Rows      []Row
	Removals  []Row
	HasChange bool
}

// Build computes install/upgrade/unchanged/remove decisions.
func Build(in Input, opts Options) Result {
	wantByKey := map[string]Wanted{}
	for _, w := range in.Wanted {
		wantByKey[key(w.Name, w.Namespace)] = w
	}
	obsByKey := map[string]Observed{}
	for _, o := range in.Observed {
		obsByKey[key(o.Name, o.Namespace)] = o
	}

	var res Result
	seen := map[string]bool{}

	for _, w := range in.Wanted {
		k := key(w.Name, w.Namespace)
		if opts.Only != "" && !opts.OnlyClosure[w.Name] {
			continue
		}
		o, live := obsByKey[k]
		seen[k] = true
		row := Row{Name: w.Name, Namespace: w.Namespace}
		if !live {
			row.Action = ActionInstall
			row.ToVersion = w.Version
			res.Rows = append(res.Rows, row)
			res.HasChange = true
			continue
		}
		if isStuck(o.Status) {
			row.Action = ActionStuck
			row.Note = "release is pending; run: rtifact deploy rollback " + w.Name
			res.Rows = append(res.Rows, row)
			res.HasChange = true
			continue
		}
		if !o.ManagedByDeployment {
			if opts.AllowAdopt {
				row.Action = ActionAdopt
				row.ToVersion = w.Version
			} else {
				row.Action = ActionRefused
				row.Note = "release exists but is not managed by this deployment (use --adopt)"
			}
			res.Rows = append(res.Rows, row)
			res.HasChange = row.Action != ActionUnchanged
			continue
		}
		same := o.Chart == w.Chart && o.Version == w.Version && o.ValuesDigest == w.ValuesDigest
		if same {
			row.Action = ActionUnchanged
			row.FromVersion = o.Version
			row.ToVersion = w.Version
			res.Rows = append(res.Rows, row)
			continue
		}
		if versionLT(w.Version, o.Version) && !opts.AllowDowngrade {
			row.Action = ActionRefused
			row.FromVersion = o.Version
			row.ToVersion = w.Version
			row.Note = "downgrade blocked (use --allow-downgrade)"
			res.Rows = append(res.Rows, row)
			continue
		}
		row.Action = ActionUpgrade
		row.FromVersion = o.Version
		row.ToVersion = w.Version
		res.Rows = append(res.Rows, row)
		res.HasChange = true
	}

	// Removals: in inventory or cluster but not in wanted file.
	for _, o := range in.Observed {
		k := key(o.Name, o.Namespace)
		if seen[k] {
			continue
		}
		if opts.Only != "" && !opts.OnlyClosure[o.Name] {
			continue
		}
		row := Row{Name: o.Name, Namespace: o.Namespace, Action: ActionRemove, FromVersion: o.Version}
		if w := findInv(in.Inventory, o.Name, o.Namespace); w != nil && w.Protect && !opts.AllowDestroy {
			row.Action = ActionRefused
			row.Note = "protected release (use --allow-destroy)"
		}
		res.Removals = append(res.Removals, row)
		res.HasChange = true
	}

	if len(in.Wanted) == 0 && len(in.Observed) > 0 && !opts.AllowMassRemoval {
		for i := range res.Removals {
			res.Removals[i].Action = ActionRefused
			res.Removals[i].Note = "mass removal blocked (use --allow-mass-removal)"
		}
	} else if massRemoval(in, res.Removals) && !opts.AllowMassRemoval {
		for i := range res.Removals {
			if res.Removals[i].Action == ActionRemove {
				res.Removals[i].Action = ActionRefused
				res.Removals[i].Note = "mass removal blocked (use --allow-mass-removal)"
			}
		}
	}

	res.Rows = append(res.Rows, res.Removals...)
	return res
}

func key(name, ns string) string { return name + "\x00" + ns }

func findInv(inv []Wanted, name, ns string) *Wanted {
	for i := range inv {
		if inv[i].Name == name && inv[i].Namespace == ns {
			return &inv[i]
		}
	}
	return nil
}

func isStuck(status string) bool {
	switch status {
	case "pending-install", "pending-upgrade", "pending-rollback":
		return true
	default:
		return false
	}
}

func versionLT(a, b string) bool {
	// semver compare via simple string when equal lengths; fallback string compare
	if a == b {
		return false
	}
	return a < b
}

func massRemoval(in Input, removals []Row) bool {
	if len(in.Inventory) == 0 {
		return len(removals) > 0
	}
	n := 0
	for _, r := range removals {
		if r.Action == ActionRemove {
			n++
		}
	}
	return n*2 > len(in.Inventory)
}