package deploy

import (
	"context"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/cluster"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
	"helm.sh/helm/v4/pkg/registry"
)

// Session holds validated deploy runtime state shared by every deploy command:
// one Kubernetes clientset and one chart cache per run.
type Session struct {
	Doc      *Document
	Disc     Discovery
	Kube     kube.Config
	Client   cluster.Client
	Charts   *helmclient.ChartCache
	Identity IdentityInfo
}

// OpenSession validates the deploy file, loads it, discovers cluster paths, and checks identity.
func OpenSession(ctx context.Context, o RunOptions) (*Session, error) {
	_ = ctx
	if err := ValidateFile(o.FilePath); err != nil {
		return nil, err
	}
	doc, err := LoadFile(o.FilePath)
	if err != nil {
		return nil, err
	}
	disc, err := Discover(DiscoverOpts{
		FilePath: o.FilePath, Doc: doc, KubeconfigCLI: o.KubeconfigCLI,
	})
	if err != nil {
		return nil, err
	}
	kcfg := kube.Config{
		Kubeconfig: disc.KubeconfigPath,
		Context:    doc.Cluster.Context,
	}
	id, err := CheckIdentity(kcfg, doc, disc)
	if err != nil {
		return nil, err
	}
	if err := validateOnlyRelease(doc, o.Only); err != nil {
		return nil, err
	}
	cs, err := cluster.New(kcfg)
	if err != nil {
		return nil, err
	}
	return &Session{
		Doc: doc, Disc: disc, Kube: kcfg, Client: cs,
		Charts: helmclient.NewChartCache(), Identity: id,
	}, nil
}

// DeploymentID is the inventory/lock key for this session.
func (s *Session) DeploymentID() string { return DeploymentID(s.Doc) }

func (s *Session) helm(o RunOptions, reg *registry.Client) (*helmclient.Client, error) {
	return helmclient.New(helmclient.Options{
		Kube: s.Kube, Deployment: s.DeploymentID(), RegistryClient: reg,
		Defaults: helmDefaults(o, s.Doc),
		Registry: helmclient.RegistryOpts{PlainHTTP: s.Doc.Registry.PlainHTTP},
		Charts:   s.Charts,
	})
}
