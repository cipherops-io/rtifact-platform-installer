package deploy

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/ocireg"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
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

func ensurePullSecrets(ctx context.Context, kcfg kube.Config, doc *Document, cred ocireg.Credentials) error {
	needsPull := false
	for _, ns := range doc.Namespaces {
		if ns.PullSecret {
			needsPull = true
			break
		}
	}
	if needsPull && doc.Registry.Auth == nil {
		return fmt.Errorf("registry.auth is required when pull_secret is enabled on a namespace")
	}
	if !needsPull {
		return nil
	}
	name := pullSecretName(doc)
	host := doc.Registry.Host
	data, err := dockerConfigJSON(host, cred)
	if err != nil {
		return err
	}
	cs, err := kube.Clientset(kcfg)
	if err != nil {
		return err
	}
	for _, ns := range doc.Namespaces {
		if !ns.PullSecret {
			continue
		}
		sec := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: ns.Name,
				Labels:    map[string]string{"rtifact.rtifact.ai/managed": "pull-secret"},
			},
			Type: corev1.SecretTypeDockerConfigJson,
			Data: map[string][]byte{
				corev1.DockerConfigJsonKey: data,
			},
		}
		existing, err := cs.CoreV1().Secrets(ns.Name).Get(ctx, name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			if _, err := cs.CoreV1().Secrets(ns.Name).Create(ctx, sec, metav1.CreateOptions{}); err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if !bytes.Equal(existing.Data[corev1.DockerConfigJsonKey], data) {
			sec.ResourceVersion = existing.ResourceVersion
			if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				cur, getErr := cs.CoreV1().Secrets(ns.Name).Get(ctx, name, metav1.GetOptions{})
				if getErr != nil {
					return getErr
				}
				if bytes.Equal(cur.Data[corev1.DockerConfigJsonKey], data) {
					return nil
				}
				upd := sec.DeepCopy()
				upd.ResourceVersion = cur.ResourceVersion
				_, updErr := cs.CoreV1().Secrets(ns.Name).Update(ctx, upd, metav1.UpdateOptions{})
				return updErr
			}); err != nil {
				return err
			}
		}
		if ns.AttachPullSecretToDefaultSA {
			if err := attachPullSecretToDefaultSA(ctx, cs, ns.Name, name); err != nil {
				return err
			}
		}
	}
	return nil
}

func attachPullSecretToDefaultSA(ctx context.Context, cs *kubernetes.Clientset, ns, secretName string) error {
	if err := wait.PollUntilContextTimeout(ctx, 500*time.Millisecond, 2*time.Minute, true, func(ctx context.Context) (bool, error) {
		_, err := cs.CoreV1().ServiceAccounts(ns).Get(ctx, "default", metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return err == nil, err
	}); err != nil {
		return fmt.Errorf("default service account in %q: %w", ns, err)
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		sa, err := cs.CoreV1().ServiceAccounts(ns).Get(ctx, "default", metav1.GetOptions{})
		if err != nil {
			return err
		}
		for _, s := range sa.ImagePullSecrets {
			if s.Name == secretName {
				return nil
			}
		}
		sa.ImagePullSecrets = append(sa.ImagePullSecrets, corev1.LocalObjectReference{Name: secretName})
		_, err = cs.CoreV1().ServiceAccounts(ns).Update(ctx, sa, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("default service account in %q: %w", ns, err)
		}
		return nil
	})
}