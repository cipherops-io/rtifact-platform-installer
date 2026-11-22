package runner

import (
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

func TestAWSdestroyGuardSkipsNonAWSLayer(t *testing.T) {
	c := &config.CustomerConfig{Cloud: "aws", Account: config.AccountBlock{Region: "us-east-1"}}
	err := awsDestroyOrphanGuard(c, stacks.LayerStateBackend, map[string]any{
		"state": map[string]any{"air_gapped": true},
	})
	if err != nil {
		t.Fatalf("unexpected guard: %v", err)
	}
}