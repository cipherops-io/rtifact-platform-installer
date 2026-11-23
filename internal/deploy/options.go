package deploy

import (
	"io"
	"os"
	"time"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
)

// Guards are the plan safety overrides; only plan and apply read them.
type Guards struct {
	Only             string
	AllowDestroy     bool
	AllowMassRemoval bool
	AllowDowngrade   bool
	AllowAdopt       bool
	Purge            bool
}

func (g Guards) planOptions(doc *Document) plan.Options {
	return plan.Options{
		AllowDestroy: g.AllowDestroy, AllowMassRemoval: g.AllowMassRemoval,
		AllowDowngrade: g.AllowDowngrade, AllowAdopt: g.AllowAdopt, Purge: g.Purge,
		Only: g.Only, OnlyClosure: onlyClosure(doc, g.Only),
	}
}

// RunOptions are the deploy CLI flags. Guards apply to plan/apply; destroy reads
// only AllowDestroy and Purge; ForceUnlock applies to every locking command.
type RunOptions struct {
	FilePath         string
	KubeconfigCLI    string
	Yes              bool
	DetailedExitCode bool
	Timeout          time.Duration
	ForceUnlock      bool
	Debug            bool
	JSONOutput       bool
	Guards
	Stdout io.Writer
	Stderr io.Writer
}

func (o *RunOptions) out() io.Writer {
	if o.Stdout != nil {
		return o.Stdout
	}
	return os.Stdout
}

func (o *RunOptions) err() io.Writer {
	if o.Stderr != nil {
		return o.Stderr
	}
	return os.Stderr
}
