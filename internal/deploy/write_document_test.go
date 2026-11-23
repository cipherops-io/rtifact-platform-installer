package deploy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteDocument_roundTripValidates(t *testing.T) {
	root := filepath.Join("..", "..")
	for _, src := range []string{
		"examples/deploy/minimal.yaml",
		"examples/deploy/demo/demo.yaml",
		"examples/deploy/reference/reference.yaml",
		"tests/deploy/fixtures/e2e.yaml",
	} {
		t.Run(src, func(t *testing.T) {
			path := filepath.Join(root, src)
			doc, err := LoadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			for _, r := range doc.Releases {
				for _, vf := range r.ValuesFiles {
					b, err := os.ReadFile(filepath.Join(filepath.Dir(path), vf))
					if err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(dir, vf), b, 0o600); err != nil {
						t.Fatal(err)
					}
				}
			}
			out := filepath.Join(dir, "releases.yaml")
			if err := WriteDocument(out, doc); err != nil {
				t.Fatal(err)
			}
			if err := ValidateFile(out); err != nil {
				b, _ := os.ReadFile(out)
				t.Fatalf("rewritten %s does not validate: %v\n%s", src, err, b)
			}
		})
	}
}
