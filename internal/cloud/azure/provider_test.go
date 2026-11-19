package azure

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud"
)

func spec() cloud.StateSpec {
	return cloud.StateSpec{
		Profile:    "ai-os-k8s",
		NamePrefix: "cm-aios-sbx",
		AirGapped:  true,
		Account:    cloud.Account{SubscriptionID: "11111111-1111-1111-1111-111111111111", TenantID: "22222222-2222-2222-2222-222222222222", Environment: "public"},
	}
}

func TestStateNames(t *testing.T) {
	if got := StateResourceGroupName("cm-aios-sbx"); got != "cm-aios-sbx-tfstate-rg" {
		t.Fatalf("rg = %q", got)
	}
	if got := StateStorageAccountName("cm-aios-sbx"); got != "cmaiossbxtfstate" {
		t.Fatalf("sa = %q", got)
	}
	long := StateStorageAccountName("averyveryverylong-workload-prefix")
	if len(long) != StorageAccountNameMaxLen || !strings.HasSuffix(long, "tfstate") {
		t.Fatalf("long sa = %q (%d)", long, len(long))
	}
}

func TestBackendRemote(t *testing.T) {
	b, err := New().Backend(spec(), "core-network")
	if err != nil {
		t.Fatal(err)
	}
	if b.Type != "azurerm" {
		t.Fatalf("type = %q", b.Type)
	}
	want := map[string]string{
		"resource_group_name":  "cm-aios-sbx-tfstate-rg",
		"storage_account_name": "cmaiossbxtfstate",
		"container_name":       "tfstate",
		"key":                  "ai-os-k8s/cm-aios-sbx/core-network.tfstate",
		"use_azuread_auth":     "true",
	}
	for k, v := range want {
		if b.Settings[k] != v {
			t.Fatalf("%s = %q, want %q", k, b.Settings[k], v)
		}
	}
	hcl := b.HCL()
	if !strings.Contains(hcl, `backend "azurerm"`) || !strings.Contains(hcl, "use_azuread_auth = true") {
		t.Fatalf("hcl:\n%s", hcl)
	}
}

func TestBackendOverrideKeyPerLayer(t *testing.T) {
	s := spec()
	s.Override = &cloud.BackendOverride{ResourceGroupName: "rg", StorageAccountName: "sa", ContainerName: "c", Key: "my.tfstate"}
	a, _ := New().Backend(s, "core-network")
	b, _ := New().Backend(s, "data-platform")
	if a.Settings["key"] == b.Settings["key"] {
		t.Fatalf("layers share key %q", a.Settings["key"])
	}
	if a.Settings["storage_account_name"] != "sa" || a.Settings["key"] != "my/core-network.tfstate" {
		t.Fatalf("override not applied: %v", a.Settings)
	}
}

func TestBackendLocal(t *testing.T) {
	s := spec()
	s.Local = true
	s.LocalDir = t.TempDir()
	b, err := cloud.Backend(New(), s, "core-network")
	if err != nil {
		t.Fatal(err)
	}
	if b.Type != "local" || !strings.HasSuffix(b.Settings["path"], "core-network.tfstate") {
		t.Fatalf("local backend = %+v", b)
	}
	rs := b.RemoteState()
	if rs["backend"] != "local" {
		t.Fatalf("remote state = %v", rs)
	}
}

func TestEnsureStateBackendNotAirGapped(t *testing.T) {
	s := spec()
	s.AirGapped = false
	_, err := New().EnsureStateBackend(context.Background(), s)
	if !errors.Is(err, cloud.ErrNotImplemented) {
		t.Fatalf("err = %v, want ErrNotImplemented", err)
	}
}

type fakeAz struct {
	outputs map[string]string
	fail    map[string]bool
	calls   []string
}

func (f *fakeAz) run(_ context.Context, name string, args ...string) ([]byte, error) {
	key := strings.Join(args[:min(2, len(args))], " ")
	if len(args) > 2 && args[0] == "storage" {
		key = strings.Join(args[:3], " ")
	}
	f.calls = append(f.calls, name+" "+key)
	if f.fail[key] {
		return nil, fmt.Errorf("exit 3")
	}
	return []byte(f.outputs[key] + "\n"), nil
}

func TestEnsureStateBackendOK(t *testing.T) {
	f := &fakeAz{outputs: map[string]string{
		"account show":             "11111111-1111-1111-1111-111111111111",
		"group exists":             "true",
		"storage account show":     "cmaiossbxtfstate",
		"storage container exists": "true",
	}}
	info, err := NewWithRunner(f.run).EnsureStateBackend(context.Background(), spec())
	if err != nil {
		t.Fatal(err)
	}
	if info.StorageAccountName != "cmaiossbxtfstate" || len(f.calls) != 4 {
		t.Fatalf("info=%+v calls=%v", info, f.calls)
	}
}

func TestEnsureStateBackendMissing(t *testing.T) {
	f := &fakeAz{outputs: map[string]string{
		"account show": "11111111-1111-1111-1111-111111111111",
		"group exists": "false",
	}}
	_, err := NewWithRunner(f.run).EnsureStateBackend(context.Background(), spec())
	if !errors.Is(err, cloud.ErrStateBackendMissing) {
		t.Fatalf("err = %v", err)
	}
}

func TestEnsureStateBackendWrongSubscription(t *testing.T) {
	f := &fakeAz{outputs: map[string]string{"account show": "99999999-9999-9999-9999-999999999999"}}
	if _, err := NewWithRunner(f.run).EnsureStateBackend(context.Background(), spec()); err == nil {
		t.Fatal("expected error")
	}
}
