package helmclient

import (
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
	"helm.sh/helm/v4/pkg/cli"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// kubeAccess builds a RESTClientGetter from explicit kubeconfig and context only (no HELM_* / KUBECONFIG env).
type kubeAccess struct {
	kubeConfig string
	context    string
	flags      *genericclioptions.ConfigFlags
}

func newKubeAccess(k kube.Config) *kubeAccess {
	ka := &kubeAccess{
		kubeConfig: k.Kubeconfig,
		context:    k.Context,
	}
	ka.flags = &genericclioptions.ConfigFlags{
		KubeConfig: &ka.kubeConfig,
		Context:    &ka.context,
	}
	return ka
}

func (ka *kubeAccess) RESTClientGetter() genericclioptions.RESTClientGetter {
	return ka.flags
}

// chartLocateEnv supplies cache paths for Helm chart download without cli.New() defaults.
func chartLocateEnv(cacheRoot string) *cli.EnvSettings {
	return &cli.EnvSettings{
		RepositoryCache:  filepath.Join(cacheRoot, "repository"),
		RepositoryConfig: filepath.Join(cacheRoot, "repositories.yaml"),
		ContentCache:     filepath.Join(cacheRoot, "content"),
	}
}