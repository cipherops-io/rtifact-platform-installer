package deploy

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	envRef   = regexp.MustCompile(`^\$\{env\.[A-Z][A-Z0-9_]*\}$`)
	infraRef = regexp.MustCompile(`^\$\{infra\.[a-zA-Z0-9_.-]+\}$`)
)

func validatePlaceholders(values map[string]any) error {
	return walkPlaceholders(values, func(s string) error {
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