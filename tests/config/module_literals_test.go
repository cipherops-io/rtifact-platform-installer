package config_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var sensitiveModuleLiteral = regexp.MustCompile(`(vm_size|node_count|sku_tier|account_tier|database_name|admin_username)\s*=\s*("[^"]+"|[0-9]+)`)

func TestModulesSensitiveLiteralsDocumentedInVars(t *testing.T) {
	root := filepath.Join("..", "..", "infra", "modules")
	var hits []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, "main.tofu") {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for i, line := range strings.Split(string(b), "\n") {
			if sensitiveModuleLiteral.MatchString(strings.TrimSpace(line)) {
				hits = append(hits, filepath.ToSlash(path)+":"+strconv.Itoa(i+1)+":"+strings.TrimSpace(line))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) > 0 {
		t.Logf("modules still using inline sensitive literals (Phase 4 A1 backlog):\n  %s", strings.Join(hits, "\n  "))
	}
}
