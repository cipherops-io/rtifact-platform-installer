package quota

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
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

func placeholderSubscription(id string) bool {
	id = strings.TrimSpace(strings.ToLower(id))
	return id == "" || id == "00000000-0000-0000-0000-000000000000"
}

func fetchUsages(ctx context.Context, subscriptionID, region string) (map[string]struct{ current, limit int }, error) {
	if placeholderSubscription(subscriptionID) {
		return nil, usererr.WithHint(
			"Cannot check Azure quota with a placeholder subscription.",
			"Set account.subscription_id to your real subscription UUID, or run validate with --check-azure=false.",
			`The sample value 00000000-0000-0000-0000-000000000000 is not valid for live quota checks.`,
		)
	}
	if strings.TrimSpace(region) == "" {
		return nil, usererr.New("Cannot check quota: account.region is empty.", "Set account.region (e.g. centralindia).")
	}
	az, err := exec.LookPath("az")
	if err != nil {
		return nil, usererr.WithHint(
			"Azure CLI (az) is not installed or not on PATH.",
			"Install Azure CLI, sign in with az login, or run validate with --check-azure=false.",
		)
	}
	_ = ctx
	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/providers/Microsoft.Compute/locations/%s/usages?api-version=2023-07-01", subscriptionID, region)
	cmd := exec.Command(az, "rest", "--method", "get", "--url", url)
	out, err := cmd.CombinedOutput()
	if err != nil {
		detail := shortenAzureCLIOutput(strings.TrimSpace(string(out)))
		if detail == "" {
			detail = err.Error()
		}
		low := strings.ToLower(detail)
		if strings.Contains(low, "operation not permitted") && strings.Contains(low, ".azure") {
			return nil, usererr.WithHint(
				"Azure CLI could not write command logs under ~/.azure.",
				"Fix permissions on your home directory, reinstall Azure CLI, or run validate with --check-azure=false.",
				detail,
			)
		}
		if strings.Contains(low, "not logged in") || strings.Contains(low, "please run 'az login'") {
			return nil, usererr.WithHint(
				"You are not signed in to Azure CLI.",
				"Run `az login`, confirm `az account show`, then retry — or use --check-azure=false.",
			)
		}
		if strings.Contains(low, "subscription") && strings.Contains(low, "not found") {
			return nil, usererr.WithHint(
				"Azure subscription was not found.",
				"Fix account.subscription_id in your YAML and ensure your login can access that subscription (`az account list`).",
				detail,
			)
		}
		return nil, usererr.WithHint(
			"Azure quota API call failed.",
			"Verify `az login`, subscription access, and region spelling — or use --check-azure=false.",
			detail,
		)
	}
	var page usagePage
	if err := json.Unmarshal(out, &page); err != nil {
		return nil, usererr.Wrap(err, "Azure quota API returned invalid JSON.", "Retry later or use --check-azure=false.")
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

func shortenAzureCLIOutput(detail string) string {
	var kept []string
	for _, line := range strings.Split(detail, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Traceback") || strings.HasPrefix(line, "File ") || strings.HasPrefix(line, "To check existing") {
			continue
		}
		if strings.HasPrefix(line, "ERROR:") {
			kept = append(kept, strings.TrimSpace(strings.TrimPrefix(line, "ERROR:")))
			continue
		}
		if !strings.Contains(line, "site-packages") && len(line) < 200 {
			kept = append(kept, line)
		}
	}
	if len(kept) > 0 {
		if len(kept) > 3 {
			kept = kept[:3]
		}
		return strings.Join(kept, " ")
	}
	if len(detail) > 320 {
		return detail[:320] + "…"
	}
	return detail
}

func poolsForFamily(pools []PoolReq, family string) []PoolReq {
	var out []PoolReq
	for _, p := range pools {
		if VMFamily(p.VMSize) == family {
			out = append(out, p)
		}
	}
	return out
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
