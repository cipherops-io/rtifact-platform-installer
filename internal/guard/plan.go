package guard

import (
	"encoding/json"
	"fmt"
)

var DefaultProtectedTypes = func() []string {
	return []string{
		"azurerm_postgresql_flexible_server",
		"azurerm_storage_account",
		"azurerm_key_vault",
		"azurerm_container_registry",
		"azurerm_kubernetes_cluster",
		"azurerm_resource_group",
		"azurerm_private_dns_zone",
	}
}

type planJSON struct {
	ResourceChanges []struct {
		Address string `json:"address"`
		Type    string `json:"type"`
		Change  struct {
			Actions []string `json:"actions"`
		} `json:"change"`
	} `json:"resource_changes"`
}

func CheckPlanJSON(b []byte, protected []string) error {
	var p planJSON
	if err := json.Unmarshal(b, &p); err != nil {
		return fmt.Errorf("plan guard: parse plan JSON: %w", err)
	}
	pro := map[string]bool{}
	for _, t := range protected {
		pro[t] = true
	}
	var hits []string
	for _, rc := range p.ResourceChanges {
		if !pro[rc.Type] {
			continue
		}
		for _, a := range rc.Change.Actions {
			if a == "delete" || a == "replace" {
				hits = append(hits, fmt.Sprintf("%s (%s)", rc.Address, a))
				break
			}
		}
	}
	if len(hits) == 0 {
		return nil
	}
	return fmt.Errorf("stateful resource changes blocked: %v (pass --allow-destroy-stateful to override)", hits)
}
