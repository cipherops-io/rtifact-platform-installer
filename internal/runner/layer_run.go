package runner

import (
	"context"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud"
	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

type layerRun struct {
	ctx         context.Context
	opts        Options
	customer    *config.CustomerConfig
	meta        *config.ProfileMeta
	provider    cloud.Cloud
	spec        cloud.StateSpec
	contentRoot string
	merged      map[string]any
	layer       stacks.Layer

	workDir     string
	stackWork   string
	tofuPath    string
	tofuEnv     []string
	tofu        TofuRunner
	varFileArg  string
	tofuCleanup func()
	planSkipped bool
}