package status

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

func Print(configPath string) error {
	cfg, err := config.LoadCustomer(configPath)
	if err != nil {
		return err
	}
	dir := filepath.Join(os.Getenv("HOME"), ".rtifact-installer", cfg.NamePrefix())
	for _, layer := range stacks.ApplyOrder {
		path := filepath.Join(dir, fmt.Sprintf("%s-outputs.json", layer))
		state := "not_started"
		if _, err := os.Stat(path); err == nil {
			state = "applied"
		}
		fmt.Printf("%s\t%s\n", layer, state)
	}
	return nil
}