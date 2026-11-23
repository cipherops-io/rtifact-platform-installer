// Package cluster holds the deploy engine's direct Kubernetes API calls
// (namespaces, pull secrets, PVC purge). Orchestration in internal/deploy
// talks to the cluster only through this package, lock, inventory, crds and helmclient.
package cluster

import (
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
	"k8s.io/client-go/kubernetes"
)

// Client is the typed Kubernetes API shared by one deploy session.
type Client = kubernetes.Interface

// New builds the session clientset from an explicit kubeconfig.
func New(cfg kube.Config) (Client, error) {
	return kube.Clientset(cfg)
}
