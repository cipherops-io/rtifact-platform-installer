package deploy

import (
	"fmt"
	"os"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"gopkg.in/yaml.v3"
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
	return walkPlaceholders(v, checkPlaceholderString)
}

// checkPlaceholderString rejects only malformed ${env.*} or ${infra.*} references.
// Chart-owned tokens such as ${DS_PROMETHEUS} are left untouched.
func checkPlaceholderString(s string) error {
	if !strings.Contains(s, "${") {
		return nil
	}
	for i := 0; i < len(s); {
		j := strings.Index(s[i:], "${")
		if j < 0 {
			break
		}
		start := i + j
		end := strings.Index(s[start:], "}")
		if end < 0 {
			break
		}
		token := s[start : start+end+1]
		switch {
		case strings.HasPrefix(token, "${env."):
			if !EnvPlaceholderToken.MatchString(token) {
				return fmt.Errorf("malformed env placeholder %q (use ${env.NAME} with uppercase NAME)", token)
			}
		case strings.HasPrefix(token, "${infra."):
			if !InfraPlaceholderToken.MatchString(token) {
				return fmt.Errorf("malformed infra placeholder %q (use ${infra.key})", token)
			}
		}
		i = start + end + 1
	}
	return nil
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