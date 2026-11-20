package config_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNoBoilerplateYmlUnderInfra(t *testing.T) {
	infra := "../../infra"
	var bad []string
	_ = filepath.WalkDir(infra, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == "boilerplate.yml" {
			bad = append(bad, p)
		}
		return nil
	})
	if len(bad) > 0 {
		t.Fatalf("boilerplate.yml must not live under infra/ (remove template copies): %v", bad)
	}
}