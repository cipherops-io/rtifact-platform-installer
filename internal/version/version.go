package version

// Set via -ldflags at release.
var (
	Version   = "0.1.0-dev"
	GitCommit = "unknown"
	Release   = "false"
)

// IsRelease reports whether the binary was built with release ldflags.
func IsRelease() bool {
	return Release == "true"
}