package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq" // PostgreSQL driver conn
	"github.com/pressly/goose"
)

type MigrationConfig struct {
	Action         string `env:"GOOSE_MIGRATE"`
	Driver         string `env:"GOOSE_DRIVER"`
	ConnStr        string `env:"GOOSE_DBSTRING"`
	MigrationsPath string `env:"GOOSE_MIGRATION_DIR"`
}

// action - up migrates up, down migrates down, everything else return nil
func Migrate(cfg *MigrationConfig) error {
	db, err := sql.Open(cfg.Driver, cfg.ConnStr)
	if err != nil {
		return fmt.Errorf("failed to open migration conn: %w", err)
	}
	defer db.Close()
	switch strings.ToLower(cfg.Action) {
	case "up":
		if err := goose.Up(db, cfg.MigrationsPath); err != nil {
			return fmt.Errorf("failed to migrate UP: %w", err)
		}
	case "down":
		if err := goose.Down(db, cfg.MigrationsPath); err != nil {
			return fmt.Errorf("failed to migrate UP: %w", err)
		}
	default:
	}
	return nil
}
