package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
	"gopkg.in/yaml.v3"
)

// WriteInitConfig copies examples/<cloud>/minimal.yaml and fills tier and customer fields.
func WriteInitConfig(cloud, tier, customer, out string) error {
	cloud = strings.ToLower(strings.TrimSpace(cloud))
	if cloud == "" {
		cloud = "azure"
	}
	root, _, err := assets.ContentRoot()
	if err != nil {
		return err
	}
	src := filepath.Join(root, "examples", cloud, "minimal.yaml")
	raw, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read example %s: %w", src, err)
	}
	var doc map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("parse example: %w", err)
	}
	doc["tier"] = tier
	if customer != "" {
		doc["customer"] = map[string]any{
			"name": customer,
			"code": sanitizeCustomerCode(customer),
		}
	}
	outBytes, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}
	return os.WriteFile(out, outBytes, 0o644)
}

func sanitizeCustomerCode(customer string) string {
	code := strings.ToLower(customer)
	code = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, code)
	if len(code) < 2 {
		code = "ex"
	}
	if len(code) > 6 {
		code = code[:6]
	}
	return code
}