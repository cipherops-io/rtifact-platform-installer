package quota

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type PoolReq struct {
	Name   string
	VMSize string
	Count  int
}

type FamilyTotal struct {
	Family   string
	Required int
	Limit    int
	Current  int
	Pass     bool
}

var vmSizeRe = regexp.MustCompile(`(?i)Standard_[A-Z]+(\d+)`)

func VCPUPerNode(vmSize string) int {
	m := vmSizeRe.FindStringSubmatch(vmSize)
	if len(m) < 2 {
		return 4
	}
	var n int
	fmt.Sscanf(m[1], "%d", &n)
	if n <= 0 {
		return 4
	}
	return n
}

func VMFamily(vmSize string) string {
	vm := strings.ToLower(vmSize)
	switch {
	case strings.Contains(vm, "ds"):
		return "standardDSv5Family"
	case strings.Contains(vm, "e"):
		return "standardESv5Family"
	case strings.Contains(vm, "f"):
		return "standardFSv2Family"
	default:
		return "standardDSv5Family"
	}
}

type usagePage struct {
	Value []struct {
		Name struct {
			Value string `json:"value"`
		} `json:"name"`
		CurrentValue int `json:"currentValue"`
		Limit        int `json:"limit"`
	} `json:"value"`
}

func fetchUsages(ctx context.Context, subscriptionID, region string) (map[string]struct{ current, limit int }, error) {
	az, err := exec.LookPath("az")
	if err != nil {
		return nil, fmt.Errorf("azure quota check requires Azure CLI (az) or set --check-azure=false")
	}
	_ = ctx
	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/providers/Microsoft.Compute/locations/%s/usages?api-version=2023-07-01", subscriptionID, region)
	cmd := exec.Command(az, "rest", "--method", "get", "--url", url)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("az rest usages: %w", err)
	}
	var page usagePage
	if err := json.Unmarshal(out, &page); err != nil {
		return nil, err
	}
	limits := map[string]struct{ current, limit int }{}
	for _, u := range page.Value {
		limits[u.Name.Value] = struct{ current, limit int }{u.CurrentValue, u.Limit}
	}
	return limits, nil
}

func CheckPools(ctx context.Context, subscriptionID, region string, pools []PoolReq) ([]FamilyTotal, error) {
	need := map[string]int{}
	for _, p := range pools {
		if p.Count <= 0 || p.VMSize == "" {
			continue
		}
		fam := VMFamily(p.VMSize)
		need[fam] += VCPUPerNode(p.VMSize) * p.Count
	}
	if len(need) == 0 {
		return nil, nil
	}
	limits, err := fetchUsages(ctx, subscriptionID, region)
	if err != nil {
		return nil, err
	}
	var out []FamilyTotal
	for fam, req := range need {
		usage, ok := limits[fam]
		if !ok {
			out = append(out, FamilyTotal{Family: fam, Required: req, Pass: false})
			continue
		}
		pass := usage.limit == 0 || usage.current+req <= usage.limit
		out = append(out, FamilyTotal{
			Family: fam, Required: req, Limit: usage.limit, Current: usage.current, Pass: pass,
		})
	}
	return out, nil
}

func PrintTable(totals []FamilyTotal) {
	w := os.Stderr
	fmt.Fprintln(w, "Quota check (vCPU families):")
	for _, t := range totals {
		status := "PASS"
		if !t.Pass {
			status = "FAIL"
		}
		fmt.Fprintf(w, "  %s required=%d current=%d limit=%d %s\n", t.Family, t.Required, t.Current, t.Limit, status)
	}
}
