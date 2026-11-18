package bootstrap

import (
	"github.com/cipherops-io/rtifact-platform-installer/internal/runner"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

func Run(configPath string, yes bool) error {
	return runner.Run(runner.Options{
		Action:     "apply",
		ConfigPath: configPath,
		Stack:      stacks.LayerBootstrap,
		Yes:        yes,
	})
}