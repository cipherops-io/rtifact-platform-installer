package cluster

import (
	"bytes"
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

// PullSecretTarget is one namespace that receives the registry pull secret.
type PullSecretTarget struct {
	Namespace       string
	AttachDefaultSA bool
}

// EnsurePullSecrets creates or refreshes a dockerconfigjson secret in each target namespace.
func EnsurePullSecrets(ctx context.Context, cs Client, name string, dockerConfig []byte, targets []PullSecretTarget) error {
	for _, t := range targets {
		if err := ensurePullSecret(ctx, cs, t.Namespace, name, dockerConfig); err != nil {
			return err
		}
		if t.AttachDefaultSA {
			if err := attachPullSecretToDefaultSA(ctx, cs, t.Namespace, name); err != nil {
				return err
			}
		}
	}
	return nil
}

func ensurePullSecret(ctx context.Context, cs Client, ns, name string, data []byte) error {
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    map[string]string{"rtifact.rtifact.ai/managed": "pull-secret"},
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{corev1.DockerConfigJsonKey: data},
	}
	existing, err := cs.CoreV1().Secrets(ns).Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = cs.CoreV1().Secrets(ns).Create(ctx, sec, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}
	if bytes.Equal(existing.Data[corev1.DockerConfigJsonKey], data) {
		return nil
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		cur, getErr := cs.CoreV1().Secrets(ns).Get(ctx, name, metav1.GetOptions{})
		if getErr != nil {
			return getErr
		}
		if bytes.Equal(cur.Data[corev1.DockerConfigJsonKey], data) {
			return nil
		}
		upd := sec.DeepCopy()
		upd.ResourceVersion = cur.ResourceVersion
		_, updErr := cs.CoreV1().Secrets(ns).Update(ctx, upd, metav1.UpdateOptions{})
		return updErr
	})
}

func attachPullSecretToDefaultSA(ctx context.Context, cs Client, ns, secretName string) error {
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
