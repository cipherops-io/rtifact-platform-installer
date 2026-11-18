package quota

import (
	"context"
	"fmt"

	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
)

func PoolsFromMerged(merged map[string]any) []PoolReq {
	var pools []PoolReq
	kc, ok := merged["kubernetes_cluster"].(map[string]any)
	if !ok {
		return pools
	}
	np, ok := kc["node_pools"].(map[string]any)
	if !ok {
		return pools
	}
	for _, v := range np {
		pm, ok := v.(map[string]any)
		if !ok {
			continue
		}
		vm, _ := pm["vm_size"].(string)
		count := 1
		if c, ok := pm["node_count"].(int); ok {
			count = c
		} else if cf, ok := pm["node_count"].(float64); ok {
			count = int(cf)
		}
		name, _ := pm["name"].(string)
		pools = append(pools, PoolReq{Name: name, VMSize: vm, Count: count})
	}
	return pools
}

func ValidateMerged(customerPath string, merged map[string]any) error {
	cfg, err := config.LoadCustomer(customerPath)
	if err != nil {
		return err
	}
	totals, err := CheckPools(context.Background(), cfg.Azure.SubscriptionID, cfg.Azure.Region, PoolsFromMerged(merged))
	if err != nil {
		return err
	}
	PrintTable(totals)
	for _, t := range totals {
		if !t.Pass {
			return fmt.Errorf("quota insufficient for family %s", t.Family)
		}
	}
	return nil
}