package kube

import (
	"os"
	"path/filepath"
	"testing"
)

const twoContexts = `apiVersion: v1
kind: Config
clusters:
- name: a
  cluster: {server: "https://a.example.com:6443"}
- name: b
  cluster: {server: "https://b.example.com"}
users:
- name: u
  user: {token: t}
contexts:
- name: ctx-a
  context: {cluster: a, user: u}
- name: ctx-b
  context: {cluster: b, user: u}
current-context: ctx-a
`

func TestServerURL_honoursContextOverride(t *testing.T) {
	p := filepath.Join(t.TempDir(), "kubeconfig")
	if err := os.WriteFile(p, []byte(twoContexts), 0o600); err != nil {
		t.Fatal(err)
	}
	for ctx, want := range map[string]string{"": "https://a.example.com:6443", "ctx-b": "https://b.example.com"} {
		got, err := ServerURL(Config{Kubeconfig: p, Context: ctx})
		if err != nil || got != want {
			t.Errorf("context %q: %q, %v; want %q", ctx, got, err, want)
		}
	}
}

func TestServerURL_missingFile(t *testing.T) {
	if _, err := ServerURL(Config{Kubeconfig: filepath.Join(t.TempDir(), "nope")}); err == nil {
		t.Fatal("expected error for missing kubeconfig")
	}
}
