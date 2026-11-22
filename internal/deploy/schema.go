package deploy

import (
	"bytes"
	"fmt"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

func compileSchema() (*jsonschema.Schema, error) {
	raw, err := assets.ReadSchemaFile("deploy.schema.json")
	if err != nil {
		return nil, fmt.Errorf("deploy.schema.json: %w", err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("deploy.schema.json", bytes.NewReader(raw)); err != nil {
		return nil, err
	}
	return compiler.Compile("deploy.schema.json")
}