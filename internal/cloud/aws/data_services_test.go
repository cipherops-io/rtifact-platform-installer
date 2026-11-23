package aws

import (
	"strings"
	"testing"
)

func dataServicesConfig() map[string]any {
	return map[string]any{
		"features": map[string]any{"network": true, "postgresql": true, "redis": true},
		"network": map[string]any{
			"vpc_cidr":           "10.0.0.0/16",
			"availability_zones": []any{"us-east-1a", "us-east-1b"},
			"subnets": map[string]any{
				"datastore": map[string]any{"cidrs": []any{"10.0.1.0/24", "10.0.2.0/24"}},
			},
		},
		"postgresql": map[string]any{"subnet_tier": "datastore", "allocated_storage_gb": 100, "max_allocated_storage_gb": 500},
		"redis":      map[string]any{"subnet_tier": "datastore"},
	}
}

func TestValidateDataServices_ok(t *testing.T) {
	if err := New().Validate(dataServicesConfig()); err != nil {
		t.Fatal(err)
	}
}

func TestValidateDataServices_rejects(t *testing.T) {
	for name, tc := range map[string]struct {
		mutate func(map[string]any)
		want   string
	}{
		"unknown tier": {
			mutate: func(m map[string]any) { m["redis"].(map[string]any)["subnet_tier"] = "private" },
			want:   `redis.subnet_tier "private" is not a tier`,
		},
		"missing tier": {
			mutate: func(m map[string]any) { delete(m["postgresql"].(map[string]any), "subnet_tier") },
			want:   "set postgresql.subnet_tier",
		},
		"single zone tier": {
			mutate: func(m map[string]any) {
				m["features"].(map[string]any)["network"] = false
				netw := m["network"].(map[string]any)
				delete(netw, "availability_zones")
				netw["subnets"].(map[string]any)["datastore"] = map[string]any{"cidrs": []any{"10.0.1.0/24"}}
			},
			want: "must span at least two zones",
		},
		"max storage not above allocated": {
			mutate: func(m map[string]any) { m["postgresql"].(map[string]any)["max_allocated_storage_gb"] = 100.0 },
			want:   "max_allocated_storage_gb",
		},
	} {
		t.Run(name, func(t *testing.T) {
			m := dataServicesConfig()
			tc.mutate(m)
			if err := New().Validate(m); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func TestValidateDataServices_disabledSkipsTierCheck(t *testing.T) {
	m := dataServicesConfig()
	m["features"] = map[string]any{"network": true}
	m["redis"].(map[string]any)["subnet_tier"] = "nope"
	if err := New().Validate(m); err != nil {
		t.Fatal(err)
	}
}
