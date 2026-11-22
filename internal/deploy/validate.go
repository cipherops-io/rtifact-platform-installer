package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
	"gopkg.in/yaml.v3"
)

// ValidateFile runs JSON Schema plus semantic checks for a deploy config.
func ValidateFile(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return usererr.WithHint("Cannot read the deploy file.", "Check the path.", err.Error())
	}
	var data any
	if err := yaml.Unmarshal(raw, &data); err != nil {
		return usererr.Wrap(err, "Deploy file is not valid YAML.", "Fix indentation and quoting.")
	}
	data = config.NormalizeYAML(data)
	sch, err := compileSchema()
	if err != nil {
		return err
	}
	if err := sch.Validate(data); err != nil {
		return usererr.Wrap(err, "Deploy file failed JSON schema validation.", "Compare with examples/deploy/minimal.yaml and schemas/deploy.schema.json.")
	}
	doc, err := LoadFile(path)
	if err != nil {
		return err
	}
	if doc.SchemaVersion != SchemaVersion {
		return fmt.Errorf("schema_version must be %d", SchemaVersion)
	}
	if doc.Kind != Kind {
		return fmt.Errorf("kind must be %q", Kind)
	}
	if err := validateSemantic(path, doc); err != nil {
		return err
	}
	return nil
}

func validateSemantic(deployPath string, doc *Document) error {
	nsSet := make(map[string]struct{}, len(doc.Namespaces))
	for _, n := range doc.Namespaces {
		if _, dup := nsSet[n.Name]; dup {
			return fmt.Errorf("duplicate namespace %q", n.Name)
		}
		nsSet[n.Name] = struct{}{}
	}
	names := make(map[string]struct{}, len(doc.Releases))
	dir := filepath.Dir(deployPath)
	for _, rel := range doc.Releases {
		if _, ok := nsSet[rel.Namespace]; !ok {
			return usererr.WithHint(
				fmt.Sprintf("Release %q uses namespace %q, which is not listed under namespaces.", rel.Name, rel.Namespace),
				"Add the namespace under namespaces:, or fix release.namespace.",
				"",
			)
		}
		if _, dup := names[rel.Name]; dup {
			return fmt.Errorf("duplicate release name %q", rel.Name)
		}
		names[rel.Name] = struct{}{}
		for _, vf := range rel.ValuesFiles {
			p := vf
			if !filepath.IsAbs(p) {
				p = filepath.Join(dir, vf)
			}
			if _, err := os.Stat(p); err != nil {
				return usererr.WithHint(
					fmt.Sprintf("Release %q: values file not found: %s", rel.Name, vf),
					"Paths are relative to the deploy file directory.",
					err.Error(),
				)
			}
		}
		for _, req := range rel.Requires {
			if _, ok := names[req]; !ok {
				// allowed if defined later; full check after loop
			}
		}
		if err := validatePlaceholdersTree(rel.Values); err != nil {
			return fmt.Errorf("release %q values: %w", rel.Name, err)
		}
		for _, vf := range rel.ValuesFiles {
			p := vf
			if !filepath.IsAbs(p) {
				p = filepath.Join(dir, vf)
			}
			if err := validatePlaceholdersInFile(p); err != nil {
				return fmt.Errorf("release %q values file %q: %w", rel.Name, vf, err)
			}
		}
	}
	for _, rel := range doc.Releases {
		for _, req := range rel.Requires {
			if _, ok := names[req]; !ok {
				return usererr.WithHint(
					fmt.Sprintf("Release %q requires unknown release %q.", rel.Name, req),
					"List required releases in releases: with matching name fields.",
					"",
				)
			}
		}
	}
	order, err := InstallOrder(doc.Releases)
	if err != nil {
		return err
	}
	if len(order) != len(doc.Releases) {
		return fmt.Errorf("release graph: incomplete ordering")
	}
	return nil
}