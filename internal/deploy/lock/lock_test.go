package lock

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/inventory"
	coordinationv1 "k8s.io/api/coordination/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

const dep = "demo-dev"

func holderOf(t *testing.T, cs *fake.Clientset) string {
	t.Helper()
	l, err := cs.CoordinationV1().Leases(inventory.SystemNS).Get(context.Background(), LeaseName(dep), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return ""
	}
	if err != nil {
		t.Fatal(err)
	}
	return *l.Spec.HolderIdentity
}

func staleLease(holder string) *coordinationv1.Lease {
	old := metav1.NewMicroTime(time.Now().Add(-2 * leaseSeconds * time.Second))
	dur := int32(leaseSeconds)
	return &coordinationv1.Lease{
		ObjectMeta: metav1.ObjectMeta{Name: LeaseName(dep), Namespace: inventory.SystemNS},
		Spec:       coordinationv1.LeaseSpec{HolderIdentity: &holder, LeaseDurationSeconds: &dur, RenewTime: &old},
	}
}

func mustAcquire(t *testing.T, cs *fake.Clientset, holder string, force bool) *Handle {
	t.Helper()
	h, err := Acquire(context.Background(), cs, dep, holder, force)
	if err != nil {
		t.Fatalf("acquire %s: %v", holder, err)
	}
	t.Cleanup(func() { _ = h.Release(context.Background(), cs) })
	return h
}

func TestAcquire_secondHolderRefusedWithHolderName(t *testing.T) {
	cs := fake.NewSimpleClientset()
	mustAcquire(t, cs, "alice@a 1+r1", false)
	_, err := Acquire(context.Background(), cs, dep, "bob@b 2+r2", false)
	if err == nil || !strings.Contains(err.Error(), "Another deploy is already running") {
		t.Fatalf("want lock-held error, got %v", err)
	}
	if got := holderOf(t, cs); got != "alice@a 1+r1" {
		t.Fatalf("holder changed to %q", got)
	}
}

func TestAcquire_sameHolderReacquires(t *testing.T) {
	cs := fake.NewSimpleClientset()
	mustAcquire(t, cs, "alice", false)
	mustAcquire(t, cs, "alice", false)
}

func TestAcquire_takesOverExpiredLease(t *testing.T) {
	cs := fake.NewSimpleClientset(staleLease("crashed"))
	mustAcquire(t, cs, "alice", false)
	if got := holderOf(t, cs); got != "alice" {
		t.Fatalf("holder %q, want alice", got)
	}
}

func TestAcquire_forceUnlockClearsLiveLease(t *testing.T) {
	cs := fake.NewSimpleClientset()
	mustAcquire(t, cs, "stuck", false)
	mustAcquire(t, cs, "alice", true)
	if got := holderOf(t, cs); got != "alice" {
		t.Fatalf("holder %q, want alice", got)
	}
}

func TestAcquire_updateConflictIsLockHeld(t *testing.T) {
	cs := fake.NewSimpleClientset(staleLease("crashed"))
	cs.PrependReactor("update", "leases", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewConflict(coordinationv1.Resource("leases"), LeaseName(dep), errors.New("rv changed"))
	})
	_, err := Acquire(context.Background(), cs, dep, "alice", false)
	if err == nil || !strings.Contains(err.Error(), "Another deploy is already running") {
		t.Fatalf("want lock-held error on conflict, got %v", err)
	}
}

func TestRelease_deletesOwnLeaseOnly(t *testing.T) {
	cs := fake.NewSimpleClientset()
	h, err := Acquire(context.Background(), cs, dep, "alice", false)
	if err != nil {
		t.Fatal(err)
	}
	mustAcquire(t, cs, "bob", true)
	if err := h.Release(context.Background(), cs); err != nil {
		t.Fatal(err)
	}
	if got := holderOf(t, cs); got != "bob" {
		t.Fatalf("alice's release removed bob's lease (holder %q)", got)
	}
	if err := (&Handle{name: LeaseName(dep), holder: "bob"}).Release(context.Background(), cs); err != nil {
		t.Fatal(err)
	}
	if got := holderOf(t, cs); got != "" {
		t.Fatalf("lease still held by %q", got)
	}
}

func TestRenew_failureReachesCallback(t *testing.T) {
	cs := fake.NewSimpleClientset()
	h := mustAcquire(t, cs, "alice", false)
	var got error
	h.OnRenewFailure(func(err error) { got = err })
	boom := errors.New("apiserver gone")
	cs.PrependReactor("update", "leases", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, boom
	})
	err := h.renew(context.Background(), cs)
	if !errors.Is(err, boom) {
		t.Fatalf("renew err %v", err)
	}
	h.renewFailed(err)
	if !errors.Is(got, boom) {
		t.Fatalf("callback got %v", got)
	}
}
