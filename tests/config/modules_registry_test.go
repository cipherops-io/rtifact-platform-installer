package config_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
	"gopkg.in/yaml.v3"
)

type boilerplateVars struct {
	Modules map[string][]string `yaml:"Modules"`
}

var stackModuleSource = regexp.MustCompile(`source\s*=\s*"\.\./\.\./\.\./\.\./modules/([a-z]+)/([^"]+)"`)

func loadModulesRegistry(t *testing.T) boilerplateVars {
	t.Helper()
	b, err := os.ReadFile("../../boilerplate-vars.yml")
	if err != nil {
		t.Fatal(err)
	}
	var v boilerplateVars
	if err := yaml.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	if len(v.Modules) == 0 {
		t.Fatal("Modules missing in boilerplate-vars.yml")
	}
	return v
}

func moduleRegistered(mods map[string][]string, cloud, layer, moduleName string) bool {
	want := layer + "/" + moduleName
	for _, entry := range mods[cloud] {
		if entry == want {
			return true
		}
	}
	return false
}

func TestEveryModuleDirListedInRegistry(t *testing.T) {
	v := loadModulesRegistry(t)
	for cloud, registry := range v.Modules {
		for _, entry := range registry {
			parts := strings.SplitN(entry, "/", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				t.Errorf("Modules[%s] entry %q must be layer/module", cloud, entry)
			}
		}
		modRoot := filepath.Join("../../infra/modules", cloud)
		dirs, err := os.ReadDir(modRoot)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range dirs {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			found := false
			for _, m := range registry {
				parts := strings.SplitN(m, "/", 2)
				if len(parts) == 2 && parts[1] == name {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("infra/modules/%s/%s exists but is not listed in boilerplate-vars Modules", cloud, name)
			}
		}
	}
}

func TestStackModulesRegisteredForLayer(t *testing.T) {
	v := loadModulesRegistry(t)
	profile := "aios-infra"
	for _, cloud := range []string{"azure", "aws", "gcp"} {
		for _, layer := range stacks.AllLayers() {
			dir := filepath.Join("../../infra/stacks", profile, cloud, string(layer))
			if _, err := os.Stat(dir); err != nil {
				continue
			}
			_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() || !strings.HasSuffix(p, ".tofu") {
					return err
				}
				b, _ := os.ReadFile(p)
				for _, m := range stackModuleSource.FindAllStringSubmatch(string(b), -1) {
					if m[1] != cloud {
						continue
					}
					if !moduleRegistered(v.Modules, cloud, string(layer), m[2]) {
						t.Errorf("%s uses module %s but Modules has no %s/%s", p, m[2], layer, m[2])
					}
				}
				return nil
			})
		}
	}
}