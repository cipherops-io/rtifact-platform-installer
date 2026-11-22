package aws

import (
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

func TestDestroyOrphanGuardSkipsNonDestroyLayer(t *testing.T) {
	p := New()
	err := p.DestroyOrphanGuard(stacks.LayerStateBackend, "us-east-1", "pfx", map[string]any{
		"state": map[string]any{"air_gapped": true},
	})
	if err != nil {
		t.Fatalf("unexpected guard: %v", err)
	}
}