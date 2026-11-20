package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

// ValidateCustomerYAML checks customer config against JSON Schema.
func ValidateCustomerYAML(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return usererr.WithHint(
			"Cannot read the config file.",
			"Check the file path and permissions.",
			fmt.Sprintf("%s: %v", path, err),
		)
	}
	var data any
	if err := yaml.Unmarshal(raw, &data); err != nil {
		return usererr.Wrap(err, "Config file is not valid YAML.", "Fix indentation, colons, and quoting in your YAML.")
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
		return usererr.Wrap(err, "Config file failed JSON schema validation.", "Compare your YAML with examples/ and the schema in schemas/customer-config.schema.json.")
	}
	cfg, err := LoadCustomer(path)
	if err != nil {
		return usererr.Wrap(err, "Config file could not be loaded.", "Fix the fields reported above (profile, cloud, account, workload).")
	}
	_ = cfg

	contentRoot, _, err := assets.ContentRoot()
	if err != nil {
		return usererr.WithHint(
			"Installer content (profiles/infra) not found.",
			"From rtifact-platform-installer, run with INSTALLER_CONTENT_ROOT=. or use a release binary with embedded content.",
			err.Error(),
		)
	}
	profilesDir := filepath.Join(contentRoot, "profiles")
	defaults, err := LoadProfileDefaults(profilesDir, cfg.Profile)
	if err != nil {
		return fmt.Errorf("profile %q: %w", cfg.Profile, err)
	}
	meta, err := LoadProfileMeta(profilesDir, cfg.Profile)
	if err != nil {
		return fmt.Errorf("profile %q: %w", cfg.Profile, err)
	}
	if err := checkProfileCloud(meta, cfg.Cloud); err != nil {
		return fmt.Errorf("profile %q: %w", cfg.Profile, err)
	}
	if _, err := MergeBuildFromPath(defaults, path); err != nil {
		return err
	}
	return nil
}

// checkProfileCloud verifies the profile supports cloud (no clouds listed = any).
func checkProfileCloud(meta *ProfileMeta, cloudName string) error {
	if len(meta.Clouds) == 0 {
		return nil
	}
	for _, c := range meta.Clouds {
		if c == cloudName {
			return nil
		}
	}
	return fmt.Errorf("cloud %q is not supported (supported: %v)", cloudName, meta.Clouds)
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
