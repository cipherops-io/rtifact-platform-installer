package version

import "strings"

// PackagedClouds is injected at link time (defaults to all clouds in dev builds).
var PackagedClouds = "azure,aws,gcp"

// PackagedCloudsList returns clouds compiled into this binary.
func PackagedCloudsList() []string {
	raw := strings.TrimSpace(PackagedClouds)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// CloudPackaged reports whether cloud is embedded in this binary.
func CloudPackaged(name string) bool {
	for _, c := range PackagedCloudsList() {
		if c == name {
			return true
		}
	}
	return false
}

// OperationalClouds lists clouds with full bootstrap/apply support in this build.
func OperationalClouds() []string {
	return PackagedCloudsList()
}