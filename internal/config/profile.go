package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
	"gopkg.in/yaml.v3"
)

var (
	aliasOnce  sync.Once
	aliasIndex map[string]string
	aliasErr   error
)

type profileYAML struct {
	StackID string   `yaml:"stack_id"`
	Aliases []string `yaml:"aliases"`
}

// ResolveProfileID maps profile aliases to the canonical profiles/<id> directory name.
func ResolveProfileID(profileID string) string {
	if profileID == "" {
		return profileID
	}
	aliases := profileAliases()
	if p, ok := aliases[profileID]; ok {
		return p
	}
	return profileID
}

func profileAliases() map[string]string {
	aliasOnce.Do(func() {
		aliasIndex, aliasErr = loadProfileAliases()
	})
	if aliasErr != nil || aliasIndex == nil {
		return map[string]string{}
	}
	return aliasIndex
}

func loadProfileAliases() (map[string]string, error) {
	root, _, err := assets.ContentRootOptional()
	if err != nil {
		return nil, err
	}
	profilesDir := filepath.Join(root, "profiles")
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string)
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		id := ent.Name()
		path := filepath.Join(profilesDir, id, "profile.yaml")
		raw, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		var meta profileYAML
		if err := yaml.Unmarshal(raw, &meta); err != nil {
			return nil, err
		}
		out[id] = id
		if meta.StackID != "" && meta.StackID != id {
			out[meta.StackID] = id
		}
		for _, alias := range meta.Aliases {
			if alias != "" {
				out[alias] = id
			}
		}
	}
	return out, nil
}