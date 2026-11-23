package deploy

import (
	"sort"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
)

func inventoryFromMap(m map[string]inventory.Entry) []inventory.Entry {
	if len(m) == 0 {
		return nil
	}
	out := make([]inventory.Entry, 0, len(m))
	for _, e := range m {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Namespace != out[j].Namespace {
			return out[i].Namespace < out[j].Namespace
		}
		return out[i].Name < out[j].Name
	})
	return out
}