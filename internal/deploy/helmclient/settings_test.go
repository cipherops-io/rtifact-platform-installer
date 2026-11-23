package helmclient

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
)

const testKubeconfig = `apiVersion: v1
kind: Config
clusters:
- name: c
  cluster: {server: "https://127.0.0.1:6443"}
users:
- name: u
  user: {token: t}
contexts:
- name: ctx
  context: {cluster: c, user: u, namespace: kubeconfig-ns}
current-context: ctx
`

func TestRESTClientGetter_scopedToReleaseNamespace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kubeconfig")
	if err := os.WriteFile(path, []byte(testKubeconfig), 0o600); err != nil {
		t.Fatal(err)
	}
	ka := newKubeAccess(kube.Config{Kubeconfig: path})
	for _, want := range []string{"ns-a", "ns-b"} {
		got, _, err := ka.RESTClientGetter(want).ToRawKubeConfigLoader().Namespace()
		if err != nil || got != want {
			t.Fatalf("getter namespace = %q, %v; want %q", got, err, want)
		}
	}
}
