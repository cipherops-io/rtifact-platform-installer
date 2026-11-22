package checklist

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
)

func Render(checklistID string) (string, error) {
	if checklistID == "" {
		return "", fmt.Errorf("unknown checklist %q", checklistID)
	}
	root, _, err := assets.ContentRootOptional()
	if err != nil {
		return "", fmt.Errorf("locate checklist: %w", err)
	}
	path := filepath.Join(root, "profiles", checklistID, "checklist.md")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("unknown checklist %q", checklistID)
		}
		return "", fmt.Errorf("read checklist %q: %w", checklistID, err)
	}
	return "\n--- Post-apply checklist ---\n" + string(data), nil
}