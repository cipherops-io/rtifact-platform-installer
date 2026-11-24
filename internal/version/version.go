package version

import "runtime"

// ProductName names per-user state directories (~/.rtifact).
const ProductName = "rtifact"

// Set via -ldflags at release.
var (
	Version     = "0.0.1-dev"
	GitCommit   = "unknown"
	TofuVersion = ""
	Release     = "false"
)

// IsRelease reports whether the binary was built with release ldflags.
func IsRelease() bool {
	return Release == "true"
}

// Platform is the GOOS/GOARCH this binary runs on.
func Platform() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}
