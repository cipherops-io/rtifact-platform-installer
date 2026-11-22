package deploy

import (
	"os"

	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
	"gopkg.in/yaml.v3"
)

func LoadFile(path string) (*Document, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, usererr.WithHint(
			"Cannot read the deploy file.",
			"Check the path and permissions.",
			err.Error(),
		)
	}
	var doc Document
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, usererr.Wrap(err, "Deploy file is not valid YAML.", "Fix indentation and quoting.")
	}
	return &doc, nil
}