package lock

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
	coordinationv1 "k8s.io/api/coordination/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	coordclient "k8s.io/client-go/kubernetes/typed/coordination/v1"
)

const leaseSeconds = 3600

// Handle is an exclusive deploy lock.
type Handle struct {
	name           string
	holder         string
	cancelMu       sync.Mutex
	stopRenew      context.CancelFunc
	onRenewFailure func(error)
}

// HolderIdentity formats user@host pid+runID for the lease holder.
func HolderIdentity(runID string) string {
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	if user == "" {
		user = "unknown"
	}
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "unknown"
	}
	return fmt.Sprintf("%s@%s %d+%s", user, host, os.Getpid(), runID)
}

// OnRenewFailure runs when lease renewal fails (fail closed).
func (h *Handle) OnRenewFailure(fn func(error)) {
	h.cancelMu.Lock()
	defer h.cancelMu.Unlock()
	h.onRenewFailure = fn
}

func (h *Handle) renewFailed(err error) {
	h.cancelMu.Lock()
	fn := h.onRenewFailure
	h.cancelMu.Unlock()
	if fn != nil {
		fn(err)
		return
	}
	fmt.Fprintf(os.Stderr, "deploy lock renewal failed: %v\n", err)
}

// Acquire takes the deployment Lease; second holder is refused (S12).
func Acquire(ctx context.Context, cs kubernetes.Interface, deployment, holder string, force bool) (*Handle, error) {
	if holder == "" {
		holder = "unknown"
	}
	name := "rtifact-deploy-lock-" + deployment
	leases := cs.CoordinationV1().Leases(inventory.SystemNS)
	existing, err := leases.Get(ctx, name, metav1.GetOptions{})
	switch {
	case apierrors.IsNotFound(err):
		existing = nil
	case err != nil:
		return nil, err
	}
	if force && existing != nil && existing.Spec.HolderIdentity != nil {
		fmt.Fprintf(os.Stderr, "Force-unlock: clearing lock held by %q\n", *existing.Spec.HolderIdentity)
		_ = leases.Delete(ctx, name, metav1.DeleteOptions{})
		existing = nil
	}
	if err := checkHolder(existing, holder); err != nil {
		return nil, err
	}
	if err := writeLease(ctx, leases, newLease(name, deployment, holder), existing); err != nil {
		return nil, err
	}
	h := &Handle{name: name, holder: holder}
	renewCtx, cancel := context.WithCancel(ctx)
	h.stopRenew = cancel
	go h.renewLoop(renewCtx, cs)
	return h, nil
}

func checkHolder(existing *coordinationv1.Lease, holder string) error {
	if existing == nil || existing.Spec.HolderIdentity == nil {
		return nil
	}
	cur := *existing.Spec.HolderIdentity
	if cur == holder {
		return nil
	}
	if !expired(existing) {
		return lockHeld(fmt.Sprintf("Lock held by %q in namespace %q.", cur, inventory.SystemNS))
	}
	fmt.Fprintf(os.Stderr, "Taking over expired deploy lock held by %q (lease older than %ds).\n", cur, leaseSeconds)
	return nil
}

func newLease(name, deployment, holder string) *coordinationv1.Lease {
	now := metav1.NewMicroTime(time.Now())
	dur := int32(leaseSeconds)
	return &coordinationv1.Lease{
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
}

// writeLease creates the lease, or replaces existing only if it is unchanged since it was read.
func writeLease(ctx context.Context, leases coordclient.LeaseInterface, lease, existing *coordinationv1.Lease) error {
	if existing == nil {
		_, err := leases.Create(ctx, lease, metav1.CreateOptions{})
		if apierrors.IsAlreadyExists(err) {
			return lockHeld("The deploy lock was taken by another process while starting.")
		}
		return err
	}
	lease.ResourceVersion = existing.ResourceVersion
	_, err := leases.Update(ctx, lease, metav1.UpdateOptions{})
	if apierrors.IsConflict(err) {
		return lockHeld("The deploy lock changed while this run was taking it.")
	}
	return err
}

func lockHeld(detail string) error {
	return usererr.New(
		"Another deploy is already running.",
		detail,
		"Wait for the other run to finish or use --force-unlock if it is stale.",
	)
}

func (h *Handle) renewLoop(ctx context.Context, cs kubernetes.Interface) {
	interval := time.Duration(leaseSeconds/6) * time.Second
	if interval < 10*time.Second {
		interval = 10 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := h.renew(ctx, cs); err != nil {
				h.renewFailed(err)
				return
			}
		}
	}
}

func (h *Handle) renew(ctx context.Context, cs kubernetes.Interface) error {
	leaseClient := cs.CoordinationV1().Leases(inventory.SystemNS)
	existing, err := leaseClient.Get(ctx, h.name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if existing.Spec.HolderIdentity == nil || *existing.Spec.HolderIdentity != h.holder {
		return nil
	}
	now := metav1.NewMicroTime(time.Now())
	existing.Spec.RenewTime = &now
	_, err = leaseClient.Update(ctx, existing, metav1.UpdateOptions{})
	return err
}

func expired(l *coordinationv1.Lease) bool {
	if l.Spec.RenewTime == nil || l.Spec.LeaseDurationSeconds == nil {
		return true
	}
	until := l.Spec.RenewTime.Time.Add(time.Duration(*l.Spec.LeaseDurationSeconds) * time.Second)
	return time.Now().After(until)
}

// Release drops the lease when this process holds it.
func (h *Handle) Release(ctx context.Context, cs kubernetes.Interface) error {
	h.cancelMu.Lock()
	if h.stopRenew != nil {
		h.stopRenew()
		h.stopRenew = nil
	}
	h.cancelMu.Unlock()

	lease, err := cs.CoordinationV1().Leases(inventory.SystemNS).Get(ctx, h.name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if lease.Spec.HolderIdentity != nil && *lease.Spec.HolderIdentity != h.holder {
		return nil
	}
	return cs.CoordinationV1().Leases(inventory.SystemNS).Delete(ctx, h.name, metav1.DeleteOptions{})
}
