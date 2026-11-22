package deploy

import (
	"os"

	"gopkg.in/yaml.v3"
)

// WriteDocument saves a deploy file (for tests and init-config).
func WriteDocument(path string, doc *Document) error {
	b, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}