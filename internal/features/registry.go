package features

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Entry struct {
	Status      string   `yaml:"status"`
	Requires    []string `yaml:"requires"`
	Description string   `yaml:"description"`
	Reason      string   `yaml:"reason"`
}

type Registry struct {
	Features map[string]Entry `yaml:"features"`
}

func Load(path string) (*Registry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var r Registry
	if err := yaml.Unmarshal(raw, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func LoadProfileRegistry(profilesDir, profileID string) (*Registry, error) {
	return Load(filepath.Join(profilesDir, profileID, "features.yaml"))
}

func (r *Registry) ValidateEnabled(features map[string]bool) error {
	if r == nil {
		return nil
	}
	for name, on := range features {
		if !on {
			continue
		}
		ent, ok := r.Features[name]
		if !ok {
			return fmt.Errorf("features.%s: unknown feature (not in registry)", name)
		}
		switch ent.Status {
		case "unsupported":
			msg := ent.Reason
			if msg == "" {
				msg = "not implemented"
			}
			return fmt.Errorf("features.%s: unsupported: %s", name, msg)
		}
		for _, req := range ent.Requires {
			if !features[req] {
				return fmt.Errorf("features.%s requires features.%s true", name, req)
			}
		}
	}
	return nil
}