package inventory

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SystemNS        = "rtifact-system"
	ConfigMapPrefix = "rtifact-deploy-"
)

// Entry is one managed release in inventory.
type Entry struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Chart        string `json:"chart"`
	Version      string `json:"version"`
	ValuesDigest string `json:"values_digest"`
	Protect      bool   `json:"protect,omitempty"`
}

// Read loads inventory from the cluster ConfigMap (missing CM → empty).
func Read(ctx context.Context, kcfg kube.Config, deployment string) ([]plan.Wanted, error) {
	cs, err := kube.Clientset(kcfg)
	if err != nil {
		return nil, err
	}
	name := ConfigMapPrefix + deployment
	cm, err := cs.CoreV1().ConfigMaps(SystemNS).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	raw := cm.Data["inventory.json"]
	if raw == "" {
		return nil, nil
	}
	var entries []Entry
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		return nil, fmt.Errorf("inventory: %w", err)
	}
	out := make([]plan.Wanted, len(entries))
	for i, e := range entries {
		out[i] = plan.Wanted{
			Name: e.Name, Namespace: e.Namespace, Chart: e.Chart,
			Version: e.Version, ValuesDigest: e.ValuesDigest, Protect: e.Protect,
		}
	}
	return out, nil
}

// Write stores the full inventory snapshot.
func Write(ctx context.Context, kcfg kube.Config, deployment string, entries []Entry) error {
	cs, err := kube.Clientset(kcfg)
	if err != nil {
		return err
	}
	b, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	name := ConfigMapPrefix + deployment
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: SystemNS,
			Labels:    map[string]string{helmclient.LabelDeployment: deployment},
		},
		Data: map[string]string{"inventory.json": string(b)},
	}
	_, err = cs.CoreV1().ConfigMaps(SystemNS).Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = cs.CoreV1().ConfigMaps(SystemNS).Create(ctx, cm, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}
	_, err = cs.CoreV1().ConfigMaps(SystemNS).Update(ctx, cm, metav1.UpdateOptions{})
	return err
}