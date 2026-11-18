package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

// ValidateCustomerYAML checks customer config against JSON Schema.
func ValidateCustomerYAML(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var data any
	if err := yaml.Unmarshal(raw, &data); err != nil {
		return fmt.Errorf("yaml: %w", err)
	}
	data = normalizeYAML(data)

	schemaRaw, err := fsReadSchema()
	if err != nil {
		return err
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("customer-config.schema.json", bytes.NewReader(schemaRaw)); err != nil {
		return err
	}
	sch, err := compiler.Compile("customer-config.schema.json")
	if err != nil {
		return err
	}
	if err := sch.Validate(data); err != nil {
		return fmt.Errorf("schema validation: %w", err)
	}
	cfg, err := LoadCustomer(path)
	if err != nil {
		return err
	}
	_ = cfg

	contentRoot, _, err := assets.ContentRoot()
	if err != nil {
		return fmt.Errorf("locate profiles/infra: %w (set INSTALLER_CONTENT_ROOT or run from repo root)", err)
	}
	profilesDir := filepath.Join(contentRoot, "profiles")
	defaults, err := LoadProfileDefaults(profilesDir, cfg.Profile)
	if err != nil {
		return fmt.Errorf("profile %q: %w", cfg.Profile, err)
	}
	if _, err := MergeBuildFromPath(defaults, path); err != nil {
		return err
	}
	return nil
}

func fsReadSchema() ([]byte, error) {
	return assets.ReadSchemaFile("customer-config.schema.json")
}

func normalizeYAML(v any) any {
	switch t := v.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]any, len(t))
		for k, val := range t {
			m[fmt.Sprint(k)] = normalizeYAML(val)
		}
		return m
	case map[string]any:
		m := make(map[string]any, len(t))
		for k, val := range t {
			m[k] = normalizeYAML(val)
		}
		return m
	case []interface{}:
		s := make([]any, len(t))
		for i, val := range t {
			s[i] = normalizeYAML(val)
		}
		return s
	default:
		return v
	}
}