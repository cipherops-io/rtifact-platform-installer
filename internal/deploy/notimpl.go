package deploy

import "github.com/cipherops-io/rtifact-platform-installer/internal/usererr"

// NotImplemented reports plan/apply/status/destroy until C2+.
func NotImplemented(cmd string) error {
	return usererr.WithHint(
		cmd+" is not implemented yet.",
		"Use deploy validate and init-config; plan/apply ship in a later release step.",
		"",
	)
}