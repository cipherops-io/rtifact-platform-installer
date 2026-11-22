package config

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud/registry"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

func compileCustomerSchema(cloud string) (*jsonschema.Schema, error) {
	if cloud == "" {
		return nil, fmt.Errorf("config: cloud is required for schema validation")
	}
	name := cloud + ".schema.json"
	cloudRaw, err := assets.ReadSchemaFile(name)
	if err != nil {
		return nil, fmt.Errorf("schema for cloud %q: %w (missing schemas/%s)", cloud, err, name)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(name, bytes.NewReader(cloudRaw)); err != nil {
		return nil, err
	}
	c, err := registry.Get(cloud)
	if err != nil {
		return nil, err
	}
	if !c.CustomerSchemaUsesCommonAllOf() {
		return compiler.Compile(name)
	}
	commonRaw, err := assets.ReadSchemaFile("common.schema.json")
	if err != nil {
		return nil, fmt.Errorf("common.schema.json: %w", err)
	}
	if err := compiler.AddResource("common.schema.json", bytes.NewReader(commonRaw)); err != nil {
		return nil, err
	}
	compiled := map[string]any{
		"$id":                     "customer-config.compiled.json",
		"$schema":                 "https://json-schema.org/draft/2020-12/schema",
		"allOf":                   []map[string]string{{"$ref": "common.schema.json"}, {"$ref": name}},
		"unevaluatedProperties":   false,
	}
	b, err := json.Marshal(compiled)
	if err != nil {
		return nil, err
	}
	if err := compiler.AddResource("customer-config.compiled.json", bytes.NewReader(b)); err != nil {
		return nil, err
	}
	return compiler.Compile("customer-config.compiled.json")
}