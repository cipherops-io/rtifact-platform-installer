package kube

import (
	"fmt"
	"os"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Config selects cluster access (never ambient KUBECONFIG alone).
type Config struct {
	Kubeconfig string
	Context    string
}

// ServerURL reads the API server URL from kubeconfig without contacting the cluster.
func ServerURL(cfg Config) (string, error) {
	raw, err := os.ReadFile(cfg.Kubeconfig)
	if err != nil {
		return "", fmt.Errorf("kubeconfig: %w", err)
	}
	kcfg, err := clientcmd.Load(raw)
	if err != nil {
		return "", fmt.Errorf("kubeconfig parse: %w", err)
	}
	overrides := &clientcmd.ConfigOverrides{}
	if cfg.Context != "" {
		overrides.CurrentContext = cfg.Context
	}
	clientCfg := clientcmd.NewDefaultClientConfig(*kcfg, overrides)
	restCfg, err := clientCfg.ClientConfig()
	if err != nil {
		return "", err
	}
	return restCfg.Host, nil
}

// RESTConfig builds client REST config from an explicit kubeconfig path.
func RESTConfig(cfg Config) (*rest.Config, error) {
	raw, err := os.ReadFile(cfg.Kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("kubeconfig: %w", err)
	}
	kcfg, err := clientcmd.Load(raw)
	if err != nil {
		return nil, fmt.Errorf("kubeconfig parse: %w", err)
	}
	overrides := &clientcmd.ConfigOverrides{}
	if cfg.Context != "" {
		overrides.CurrentContext = cfg.Context
	}
	return clientcmd.NewDefaultClientConfig(*kcfg, overrides).ClientConfig()
}

// Clientset returns a Kubernetes clientset.
func Clientset(cfg Config) (*kubernetes.Clientset, error) {
	restCfg, err := RESTConfig(cfg)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(restCfg)
}

// DynamicClient returns a dynamic client for CRD inspection in e2e.
func DynamicClient(cfg Config) (dynamic.Interface, error) {
	restCfg, err := RESTConfig(cfg)
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(restCfg)
}