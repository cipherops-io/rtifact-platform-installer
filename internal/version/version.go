package version

// ProductName names per-user state directories (~/.rtifact).
const ProductName = "rtifact"

// Set via -ldflags at release.
var (
	Version   = "0.0.1-dev"
	GitCommit = "unknown"
	Customer  = "internal"
	Release   = "false"
)

// IsRelease reports whether the binary was built with release ldflags.
func IsRelease() bool {
	return Release == "true"
}
