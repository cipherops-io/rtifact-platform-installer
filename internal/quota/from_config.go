package quota

import (
	"context"
	"fmt"

	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
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
	if cfg.Cloud != "azure" {
		return nil
	}
	pools := PoolsFromMerged(merged)
	totals, err := CheckPools(context.Background(), cfg.Account.SubscriptionID, cfg.Account.Region, pools)
	if err != nil {
		return err
	}
	PrintTable(totals)
	for _, t := range totals {
		if !t.Pass {
			return quotaInsufficientError(cfg.Account.Region, t, poolsForFamily(pools, t.Family))
		}
	}
	return nil
}

func quotaInsufficientError(region string, t FamilyTotal, pools []PoolReq) error {
	details := []string{
		fmt.Sprintf("Region %s, family %s: plan needs %d vCPU; subscription shows %d used, limit %d.",
			region, t.Family, t.Required, t.Current, t.Limit),
	}
	if t.Limit == 0 && t.Current == 0 {
		details = append(details, "Azure returned no quota entry for this family (often means 0 available or wrong region/SKU family).")
	}
	for _, p := range pools {
		vcpu := VCPUPerNode(p.VMSize) * p.Count
		details = append(details, fmt.Sprintf(`Node pool %q: %s × %d node(s) → %d vCPU in this family.`, p.Name, p.VMSize, p.Count, vcpu))
	}
	return usererr.WithHint(
		"Subscription vCPU quota is too low for this config.",
		"Reduce kubernetes_cluster.node_pools (smaller vm_size or node_count), request a quota increase in Azure Portal, or run validate with --check-quota=false.",
		details...,
	)
}
