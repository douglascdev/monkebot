package monkebot

import (
	"database/sql"
	"fmt"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type Migration struct {
	Version int
	Stmts   []string
}

func InitDB(driver string, dataSourceName string, configPath string) (*sql.DB, error) {
	db, err := sql.Open(driver, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	migrations := []Migration{}
	err = RunMigrations(db, configPath, migrations)
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func CurrentSchemaDDL() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS users (
			user_id TEXT NOT NULL PRIMARY KEY,
		)`,
	}
}

// Run migrations on the database.
// If the migration succeeds, the version in DBConfig is updated to the current version.
func RunMigrations(db *sql.DB, configPath string, migrations []Migration) error {
	config, err := LoadConfigFromFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	currentVersion := config.DBConfig.Version
	// tables were not created yet, we can skip migrations and run the current DDL
	if currentVersion == 0 {
		var tx *sql.Tx
		tx, err = db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback()

		for _, stmt := range CurrentSchemaDDL() {
			_, err = tx.Exec(stmt)
			if err != nil {
				return fmt.Errorf("failed to execute DDL: %w", err)
			}
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// migrations to be applied sequentially from 0 to currentVersion
	// keep sorted by version
	for _, migration := range migrations {
		if currentVersion < migration.Version {
			break
		}

		for _, stmt := range migration.Stmts {
			_, err = tx.Exec(stmt)
			if err != nil {
				return fmt.Errorf("failed to execute migration: %w", err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	config.DBConfig.Version = currentVersion
	err = SaveConfigToFile(config, configPath)
	if err != nil {
		return fmt.Errorf("failed to update schema version to %d, please do so manually in the config file: %w", currentVersion, err)
	}
	return nil
}
