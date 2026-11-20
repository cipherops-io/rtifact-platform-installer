package testutil

import "os"

// IACBinary returns the IaC CLI (terraform or tofu) for Terratest.
func IACBinary() string {
	if v := os.Getenv("IAC_BIN"); v != "" {
		return v
	}
	return "tofu"
}