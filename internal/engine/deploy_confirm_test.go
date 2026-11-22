package engine

import (
	"testing"
)

func TestConfirmDeployDestroySkipsWithYes(t *testing.T) {
	if err := ConfirmDeployDestroy(true, "demo"); err != nil {
		t.Fatalf("got %v", err)
	}
}

func TestConfirmDeployApplySkipsWithYes(t *testing.T) {
	if err := ConfirmDeployApply(true); err != nil {
		t.Fatalf("got %v", err)
	}
	if err := ConfirmDeployRemovals(true); err != nil {
		t.Fatalf("got %v", err)
	}
}