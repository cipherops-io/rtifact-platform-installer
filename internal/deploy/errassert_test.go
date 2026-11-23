package deploy

import (
	"strings"
	"testing"
)

// wantErrContaining fails unless err is non-nil and its message mentions any of subs.
func wantErrContaining(t *testing.T, err error, subs ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error mentioning %q, got nil", subs)
	}
	msg := err.Error()
	for _, s := range subs {
		if strings.Contains(msg, s) {
			return
		}
	}
	t.Fatalf("expected error mentioning %q, got %v", subs, err)
}
