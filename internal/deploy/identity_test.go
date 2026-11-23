package deploy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeServer(t *testing.T) {
	cases := map[string]string{
		" https://a/ ":                 "https://a",
		"https://EXAMPLE.COM:443":      "https://example.com",
		"HTTPS://example.com":          "https://example.com",
		"http://Host:80":               "http://host",
		"https://api.example.com:6443": "https://api.example.com:6443",
		"https://127.0.0.1:6443/":      "https://127.0.0.1:6443",
		"https://[::1]:6443":           "https://[::1]:6443",
		"https://[::1]:443":            "https://[::1]",
		"not a url":                    "not a url",
	}
	for in, want := range cases {
		if got := normalizeServer(in); got != want {
			t.Errorf("normalizeServer(%q) = %q, want %q", in, got, want)
		}
	}
	if normalizeServer("https://Kind:443") != normalizeServer("https://kind/") {
		t.Fatal("equivalent URLs must compare equal for the identity guard")
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