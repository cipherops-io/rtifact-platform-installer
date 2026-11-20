package dataops

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
)

// Backup records a PostgreSQL backup intent for operator-driven or future API automation.
func Backup(configPath, suffix string) error {
	cfg, err := config.LoadCustomer(configPath)
	if err != nil {
		return err
	}
	if suffix == "" {
		suffix = time.Now().UTC().Format("20060102-150405")
	}
	dir := config.ArtifactsDir(cfg.NamePrefix())
	rec := fmt.Sprintf("backup_id=%s\nserver=%s\n", suffix, cfg.NamePrefix())
	return os.WriteFile(filepath.Join(dir, "data-migrations.json"), []byte(rec), 0o644)
}

func Provision(configPath, serverName string) error {
	fmt.Fprintf(os.Stderr, "provision: set overrides.postgresql_flexible_server.server_name=%q then plan/apply --stack data-platform\n", serverName)
	return nil
}

func Migrate(configPath, from, to string) error {
	fmt.Fprintf(os.Stderr, "migrate: copy data from %s to %s (see docs/operators/stateful-migrations.md)\n", from, to)
	return nil
}

func Retire(configPath, serverName string, yes bool) error {
	if !yes {
		return fmt.Errorf("retire requires --yes")
	}
	fmt.Fprintf(os.Stderr, "retire: destroy only server %q via targeted data-platform stack destroy (manual)\n", serverName)
	return nil
}
