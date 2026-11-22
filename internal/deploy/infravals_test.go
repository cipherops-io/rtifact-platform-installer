package deploy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeOutputs(t *testing.T, dir, name string, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLookupInfraScalar_layerQualified(t *testing.T) {
	dir := t.TempDir()
	writeOutputs(t, dir, "app-platform-outputs.json", `{"cluster_endpoint":"https://eks.example"}`)
	got, err := LookupInfraScalar(dir, "app-platform.cluster_endpoint")
	if err != nil || got != "https://eks.example" {
		t.Fatalf("got %q err %v", got, err)
	}
}

func TestLookupInfraScalar_bareUnique(t *testing.T) {
	dir := t.TempDir()
	writeOutputs(t, dir, "network-edge-outputs.json", `{"public_dns_zone_id":"Z123"}`)
	got, err := LookupInfraScalar(dir, "public_dns_zone_id")
	if err != nil || got != "Z123" {
		t.Fatalf("got %q err %v", got, err)
	}
}

func TestLookupInfraScalar_ambiguous(t *testing.T) {
	dir := t.TempDir()
	writeOutputs(t, dir, "app-platform-outputs.json", `{"zone_id":"a"}`)
	writeOutputs(t, dir, "network-edge-outputs.json", `{"zone_id":"b"}`)
	_, err := LookupInfraScalar(dir, "zone_id")
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("expected ambiguity, got %v", err)
	}
}

func TestLookupInfraScalar_scalarOnly(t *testing.T) {
	dir := t.TempDir()
	writeOutputs(t, dir, "app-platform-outputs.json", `{"tags":{"env":"x"}}`)
	_, err := LookupInfraScalar(dir, "app-platform.tags")
	if err == nil || !strings.Contains(err.Error(), "scalar") {
		t.Fatalf("expected scalar error, got %v", err)
	}
}

func TestLookupInfraScalar_missing(t *testing.T) {
	dir := t.TempDir()
	_, err := LookupInfraScalar(dir, "missing_key")
	if err == nil {
		t.Fatal("expected error")
	}
}