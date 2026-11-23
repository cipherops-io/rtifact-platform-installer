package config_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// inlineLiteral matches `key = "string" | true | false | number`; interpolated strings are not literals.
var inlineLiteral = regexp.MustCompile(`^\s*([a-z_0-9]+)\s*=\s*("[^"$]*"|true|false|-?[0-9]+(\.[0-9]+)?)\s*(#.*)?$`)

type literalAllowEntry struct {
	Match  string `yaml:"match"`
	Reason string `yaml:"reason"`
}

type literalAllowList struct {
	Keys     map[string]string              `yaml:"keys"`
	Literals map[string][]literalAllowEntry `yaml:"literals"`
}

func loadLiteralAllowList(t *testing.T) literalAllowList {
	t.Helper()
	b, err := os.ReadFile("module_literals_allow.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var al literalAllowList
	if err := yaml.Unmarshal(b, &al); err != nil {
		t.Fatal(err)
	}
	for k, reason := range al.Keys {
		if strings.TrimSpace(reason) == "" {
			t.Fatalf("module_literals_allow.yaml: key %q needs a reason", k)
		}
	}
	for dir, entries := range al.Literals {
		for _, e := range entries {
			if strings.TrimSpace(e.Reason) == "" {
				t.Fatalf("module_literals_allow.yaml: %s %q needs a reason", dir, e.Match)
			}
		}
	}
	return al
}

func normalizeLiteral(key, value string) string {
	return key + " = " + value
}

func skipLiteralScan(name string) bool {
	switch name {
	case "variables.tofu", "outputs.tofu", "providers.tofu", "versions.tofu", "backend.tofu":
		return true
	}
	return !strings.HasSuffix(name, ".tofu")
}

// TestInfraLiteralsAreConfigurable fails when a module or stack hardcodes a value that
// customers should be able to set from YAML. Names, provider-mandated identifiers and
// protocol constants are listed in module_literals_allow.yaml with a reason.
func TestInfraLiteralsAreConfigurable(t *testing.T) {
	al := loadLiteralAllowList(t)
	infra := filepath.Join("..", "..", "infra")
	used := map[string]map[string]bool{}
	var hits []string

	for _, sub := range []string{"modules", "stacks"} {
		root := filepath.Join(infra, sub)
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if info.Name() == ".terraform" || info.Name() == "tests" || info.Name() == "gcp" {
					return filepath.SkipDir
				}
				return nil
			}
			if skipLiteralScan(info.Name()) {
				return nil
			}
			rel, err := filepath.Rel(infra, filepath.Dir(path))
			if err != nil {
				return err
			}
			dir := filepath.ToSlash(rel)
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			for i, line := range strings.Split(string(b), "\n") {
				m := inlineLiteral.FindStringSubmatch(line)
				if m == nil {
					continue
				}
				key, value := m[1], m[2]
				if _, ok := al.Keys[key]; ok {
					continue
				}
				lit := normalizeLiteral(key, value)
				allowed := false
				for _, e := range al.Literals[dir] {
					if e.Match == lit {
						allowed = true
						if used[dir] == nil {
							used[dir] = map[string]bool{}
						}
						used[dir][lit] = true
						break
					}
				}
				if !allowed {
					hits = append(hits, filepath.ToSlash(filepath.Join(dir, info.Name()))+":"+strconv.Itoa(i+1)+": "+lit)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	if len(hits) > 0 {
		sort.Strings(hits)
		t.Errorf("hardcoded values must be variables with validation (or listed in tests/config/module_literals_allow.yaml with a reason):\n  %s", strings.Join(hits, "\n  "))
	}

	var stale []string
	for dir, entries := range al.Literals {
		for _, e := range entries {
			if !used[dir][e.Match] {
				stale = append(stale, dir+": "+e.Match)
			}
		}
	}
	if len(stale) > 0 {
		sort.Strings(stale)
		t.Errorf("module_literals_allow.yaml entries no longer match any file; remove them:\n  %s", strings.Join(stale, "\n  "))
	}
}
