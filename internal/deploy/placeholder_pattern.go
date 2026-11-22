package deploy

import "regexp"

// Shared placeholder patterns (validate + mask must stay in sync).
var (
	EnvPlaceholderToken  = regexp.MustCompile(`\$\{env\.[A-Z][A-Z0-9_]*\}`)
	InfraPlaceholderToken = regexp.MustCompile(`\$\{infra\.[a-zA-Z0-9_.-]+\}`)
)