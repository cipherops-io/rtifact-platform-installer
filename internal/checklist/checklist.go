package checklist

import (
	_ "embed"
	"fmt"
)

//go:embed ai-os-aks.md
var aiOSChecklist string

var byID = map[string]string{
	"ai-os-aks": aiOSChecklist,
}

func Render(id string) (string, error) {
	text, ok := byID[id]
	if !ok {
		return "", fmt.Errorf("unknown checklist %q", id)
	}
	return "\n--- Post-apply checklist ---\n" + text, nil
}