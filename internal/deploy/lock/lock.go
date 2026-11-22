package lock

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
	coordinationv1 "k8s.io/api/coordination/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const leaseSeconds = 3600

// Handle is an exclusive deploy lock.
type Handle struct {
	name string
}

// Acquire takes the deployment Lease; second holder is refused (S12).
func Acquire(ctx context.Context, kcfg kube.Config, deployment, holder string, force bool) (*Handle, error) {
	if holder == "" {
		holder = "unknown"
	}
	cs, err := kube.Clientset(kcfg)
	if err != nil {
		return nil, err
	}
	name := "rtifact-deploy-lock-" + deployment
	leaseClient := cs.CoordinationV1().Leases(inventory.SystemNS)
	existing, err := leaseClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}
	if force && existing != nil {
		_ = leaseClient.Delete(ctx, name, metav1.DeleteOptions{})
		existing = nil
		err = apierrors.NewNotFound(coordinationv1.Resource("leases"), name)
	}
	if err == nil && existing != nil && existing.Spec.HolderIdentity != nil {
		cur := *existing.Spec.HolderIdentity
		if cur != holder && !expired(existing) {
			return nil, usererr.New(
				"Another deploy is already running.",
				fmt.Sprintf("Lock held by %q in namespace %q.", cur, inventory.SystemNS),
				"Wait for the other run to finish or use --force-unlock if it is stale.",
			)
		}
	}
	now := metav1.NewMicroTime(time.Now())
	dur := int32(leaseSeconds)
	lease := &coordinationv1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: inventory.SystemNS,
			Labels:    map[string]string{"rtifact.rtifact.ai/deployment": deployment},
		},
		Spec: coordinationv1.LeaseSpec{
			HolderIdentity:       &holder,
			LeaseDurationSeconds: &dur,
			AcquireTime:          &now,
			RenewTime:            &now,
		},
	}
	if apierrors.IsNotFound(err) {
		_, err = leaseClient.Create(ctx, lease, metav1.CreateOptions{})
	} else {
		_, err = leaseClient.Update(ctx, lease, metav1.UpdateOptions{})
	}
	if err != nil {
		return nil, err
	}
	return &Handle{name: name}, nil
}

func expired(l *coordinationv1.Lease) bool {
	if l.Spec.RenewTime == nil || l.Spec.LeaseDurationSeconds == nil {
		return true
	}
	until := l.Spec.RenewTime.Time.Add(time.Duration(*l.Spec.LeaseDurationSeconds) * time.Second)
	return time.Now().After(until)
}

// Release drops the lease when this process holds it.
func (h *Handle) Release(ctx context.Context, kcfg kube.Config, deployment string) error {
	_ = deployment
	cs, err := kube.Clientset(kcfg)
	if err != nil {
		return err
	}
	holder, _ := os.Hostname()
	lease, err := cs.CoordinationV1().Leases(inventory.SystemNS).Get(ctx, h.name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if lease.Spec.HolderIdentity != nil && *lease.Spec.HolderIdentity != holder {
		return nil
	}
	return cs.CoordinationV1().Leases(inventory.SystemNS).Delete(ctx, h.name, metav1.DeleteOptions{})
}