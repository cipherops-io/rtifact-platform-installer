package deploy

import (
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
)

// observeTargets returns the union of file releases and inventory entries to query in Helm.
func observeTargets(want []plan.Wanted, inv []inventory.Entry) []plan.Wanted {
	seen := map[string]bool{}
	var out []plan.Wanted
	add := func(name, ns string) {
		k := plan.ReleaseKey(name, ns)
		if seen[k] {
			return
		}
		seen[k] = true
		out = append(out, plan.Wanted{Name: name, Namespace: ns})
	}
	for _, w := range want {
		add(w.Name, w.Namespace)
	}
	for _, e := range inv {
		add(e.Name, e.Namespace)
	}
	return out
}