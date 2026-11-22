package cli

import (
	"errors"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy"
	"github.com/cipherops-io/rtifact-platform-installer/internal/engine"
)

func TestExitCode(t *testing.T) {
	if ExitCode(nil) != 0 {
		t.Fatal("nil")
	}
	if ExitCode(deploy.ErrInterrupted) != 130 {
		t.Fatal("interrupt")
	}
	if ExitCode(engine.ErrAborted) != 1 {
		t.Fatal("aborted")
	}
	if ExitCode(errors.New("boom")) != 1 {
		t.Fatal("generic")
	}
	if code, ok := planExitCode(&exitStatus{code: 2}); !ok || code != 2 {
		t.Fatal("plan exit")
	}
}