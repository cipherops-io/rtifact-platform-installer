package crds

import "context"

// Upgrade applies chart CRDs when crds: upgrade (full implementation in C3).
func Upgrade(ctx context.Context, mode string, chartData []byte) error {
	_ = ctx
	if mode != "upgrade" {
		return nil
	}
	_ = chartData
	return nil
}