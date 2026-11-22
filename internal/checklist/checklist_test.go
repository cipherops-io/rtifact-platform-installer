package checklist

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderFromProfiles(t *testing.T) {
	root := filepath.Join("..", "..")
	t.Setenv("INSTALLER_CONTENT_ROOT", root)

	text, err := Render("aios-infra")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "Post-apply checklist") {
		t.Fatalf("missing header: %q", text)
	}
	if !strings.Contains(text, "MongoDB Atlas") {
		t.Fatalf("expected profile checklist content")
	}
}