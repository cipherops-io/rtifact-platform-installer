package checklist

import (
	_ "embed"
	"fmt"
)

//go:embed aios-infra.md
var aiosInfraChecklist string

var byID = map[string]string{
	"aios-infra": aiosInfraChecklist,
}

func Render(id string) (string, error) {
	text, ok := byID[id]
	if !ok {
		return "", fmt.Errorf("unknown checklist %q", id)
	}
	return "\n--- Post-apply checklist ---\n" + text, nil
}
