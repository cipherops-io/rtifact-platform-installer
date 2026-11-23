package deploy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeServer(t *testing.T) {
	if normalizeServer(" https://a/ ") != "https://a" {
		t.Fatalf("normalize failed")
	}
	if normalizeServer("https://EXAMPLE.COM:443") != "https://example.com" {
		t.Fatalf("host case-fold and default https port")
	}
	if normalizeServer("http://Host:80") != "http://host" {
		t.Fatalf("default http port")
	}
}

func TestClusterEndpointFromOutputs_missingNotError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app-platform-outputs.json")
	if err := os.WriteFile(path, []byte(`{"cluster_name":"x"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := clusterEndpointFromOutputs(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestClusterEndpointFromOutputs_present(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app-platform-outputs.json")
	if err := os.WriteFile(path, []byte(`{"cluster_endpoint":"https://eks.example"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := clusterEndpointFromOutputs(dir)
	if err != nil || got != "https://eks.example" {
		t.Fatalf("got %q err=%v", got, err)
	}
}

func TestIdentityInfo_line(t *testing.T) {
	if (IdentityInfo{Verified: true, Expected: "https://a"}).Line() != "Identity:   verified (https://a)" {
		t.Fatal("verified line")
	}
	if (IdentityInfo{}).Line() != "Identity:   not verified" {
		t.Fatal("not verified line")
	}
}