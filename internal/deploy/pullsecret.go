package deploy

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/cluster"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/ocireg"
)

const defaultPullSecretName = "rtifact-registry-pull"

func pullSecretName(doc *Document) string {
	if doc.Registry.PullSecret != nil && doc.Registry.PullSecret.Name != "" {
		return doc.Registry.PullSecret.Name
	}
	return defaultPullSecretName
}

func dockerConfigJSON(host string, cred ocireg.Credentials) ([]byte, error) {
	auth := base64.StdEncoding.EncodeToString([]byte(cred.Username + ":" + cred.Password))
	cfg := map[string]any{
		"auths": map[string]any{
			host: map[string]any{
				"username": cred.Username,
				"password": cred.Password,
				"auth":     auth,
			},
		},
	}
	return json.Marshal(cfg)
}

func ensurePullSecrets(ctx context.Context, cs cluster.Client, doc *Document, cred ocireg.Credentials) error {
	var targets []cluster.PullSecretTarget
	for _, ns := range doc.Namespaces {
		if ns.PullSecret {
			targets = append(targets, cluster.PullSecretTarget{
				Namespace: ns.Name, AttachDefaultSA: ns.AttachPullSecretToDefaultSA,
			})
		}
	}
	if len(targets) == 0 {
		return nil
	}
	if doc.Registry.Auth == nil {
		return fmt.Errorf("registry.auth is required when pull_secret is enabled on a namespace")
	}
	data, err := dockerConfigJSON(doc.Registry.Host, cred)
	if err != nil {
		return err
	}
	return cluster.EnsurePullSecrets(ctx, cs, pullSecretName(doc), data, targets)
}
