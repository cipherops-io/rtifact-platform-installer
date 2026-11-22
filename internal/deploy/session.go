package deploy

import (
	"context"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
)

// Session holds validated deploy runtime state shared by plan and apply.
type Session struct {
	Doc      *Document
	Disc     Discovery
	Kube     kube.Config
	Identity IdentityInfo
}

// OpenSession validates the deploy file, loads it, discovers cluster paths, and checks identity.
// OCI registry clients are constructed separately when RunOptions.NeedsRegistry is true (plan/apply).
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
	return &Session{Doc: doc, Disc: disc, Kube: kcfg, Identity: id}, nil
}