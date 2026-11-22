package deploy

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"gopkg.in/yaml.v3"
)

var (
	envRef   = regexp.MustCompile(`^\$\{env\.[A-Z][A-Z0-9_]*\}$`)
	infraRef = regexp.MustCompile(`^\$\{infra\.[a-zA-Z0-9_.-]+\}$`)
)

func validatePlaceholdersInFile(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var data any
	if err := yaml.Unmarshal(raw, &data); err != nil {
		return fmt.Errorf("not valid YAML: %w", err)
	}
	return validatePlaceholdersTree(config.NormalizeYAML(data))
}

func validatePlaceholders(values map[string]any) error {
	return validatePlaceholdersTree(values)
}

func validatePlaceholdersTree(v any) error {
	return walkPlaceholders(v, func(s string) error {
		if envRef.MatchString(s) || infraRef.MatchString(s) {
			return nil
		}
		if strings.Contains(s, "${") {
			return fmt.Errorf("invalid placeholder %q (use ${env.NAME} or ${infra.key})", s)
		}
		return nil
	})
}

func walkPlaceholders(v any, check func(string) error) error {
	switch t := v.(type) {
	case map[string]any:
		for _, val := range t {
			if err := walkPlaceholders(val, check); err != nil {
				return err
			}
		}
	case []any:
		for _, item := range t {
			if err := walkPlaceholders(item, check); err != nil {
				return err
			}
		}
	case string:
		return check(t)
	}
	return nil
}