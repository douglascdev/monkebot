package monkebot

import (
	"database/sql"
	"fmt"
	"sort"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type DBMigration struct {
	Version int
	Stmts   []string
}

// makes migrations sortable(implement sort.Interface)
type DBMigrations struct {
	Migrations []DBMigration
}

func (m *DBMigrations) Len() int {
	return len(m.Migrations)
}

func (m *DBMigrations) Swap(i, j int) {
	m.Migrations[i], m.Migrations[j] = m.Migrations[j], m.Migrations[i]
}

func (m *DBMigrations) Less(i, j int) bool {
	return m.Migrations[i].Version < m.Migrations[j].Version
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

	migrations := DBMigrations{
		Migrations: []DBMigration{
			{Version: 1, Stmts: CurrentSchemaDDL()},
		},
	}
	err = RunMigrations(db, configPath, &migrations)
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
func RunMigrations(db *sql.DB, configPath string, migrations *DBMigrations) error {
	// sort migrations by version
	sort.Sort(migrations)

	config, err := LoadConfigFromFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// migrations to be applied sequentially until the currentVersion
	migrationsApplied := 0
	currentVersion := config.DBConfig.Version
	for _, migration := range migrations.Migrations {
		if currentVersion < migration.Version {
			continue
		}

		for _, stmt := range migration.Stmts {
			_, err = tx.Exec(stmt)
			migrationsApplied++
			if err != nil {
				return fmt.Errorf("failed to execute migration: %w", err)
			}
		}
	}

	if migrationsApplied == 0 {
		return nil
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
