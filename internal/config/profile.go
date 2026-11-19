package config

// profileAliases maps deprecated profile ids to their replacement.
var profileAliases = map[string]string{
	"ai-os-k8s": "aios-infra",
}

// ResolveProfileID maps deprecated profile names to the current id.
func ResolveProfileID(profileID string) string {
	if p, ok := profileAliases[profileID]; ok {
		return p
	}
	return profileID
}