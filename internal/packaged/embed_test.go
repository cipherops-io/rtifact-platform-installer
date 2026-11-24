package packaged

import (
	"io/fs"
	"strings"
	"testing"
)

// Every path the CLI reads from the content root must be in the embedded tree.
func TestEmbeddedLayout(t *testing.T) {
	src := FS()
	if entries, _ := fs.ReadDir(src, "profiles"); len(entries) == 0 {
		t.Skip("no packaged content; run make package-content")
	}
	required := []string{
		"profiles",
		"infra/stacks",
		"infra/modules",
		"schemas/common.schema.json",
		"schemas/deploy.schema.json",
		"examples/deploy/minimal.yaml",
		"build/embed",
	}
	schemas, err := fs.ReadDir(src, "schemas")
	if err != nil {
		t.Fatal(err)
	}
	var clouds []string
	for _, e := range schemas {
		c := strings.TrimSuffix(e.Name(), ".schema.json")
		if c == e.Name() || c == "common" || c == "deploy" {
			continue
		}
		clouds = append(clouds, c)
		required = append(required, "examples/"+c+"/minimal.yaml", "infra/modules/"+c)
	}
	if len(clouds) == 0 {
		t.Fatal("no cloud schemas embedded")
	}
	for _, p := range required {
		if _, err := fs.Stat(src, p); err != nil {
			t.Errorf("embedded content is missing %s", p)
		}
	}
	profiles, _ := fs.ReadDir(src, "profiles")
	stacks, _ := fs.ReadDir(src, "infra/stacks")
	for _, p := range profiles {
		if _, err := fs.Stat(src, "profiles/"+p.Name()+"/profile.yaml"); err != nil {
			t.Errorf("embedded profile %s has no profile.yaml", p.Name())
		}
	}
	for _, s := range stacks {
		for _, c := range clouds {
			if _, err := fs.Stat(src, "infra/stacks/"+s.Name()+"/"+c); err != nil {
				t.Errorf("embedded stack %s has no %s layers", s.Name(), c)
			}
		}
	}
}
