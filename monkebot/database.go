package monkebot

import (
	"database/sql"
	"fmt"
	"io"
	"sort"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type DBMigration struct {
	Version int
	Stmts   []string
}

// makes migrations sortable by version(implements sort.Interface)
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

// Initialize the database, run needed migrations and update database config to the latest version if the miggrations succeed
func InitDB(driver string, dataSourceName string, cfgReader io.Reader, cfgWriter io.Writer) (*sql.DB, error) {
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
			{Version: 1, Stmts: CurrentSchema()},
		},
	}

	var cfg *Config
	var data []byte
	data, err = io.ReadAll(cfgReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg, err = LoadConfig(data)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	err = RunMigrations(tx, cfg, &migrations)
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	latestVer := migrations.Migrations[len(migrations.Migrations)-1].Version
	data, err = MarshalConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config, update your config to %d manually. Error: %w", latestVer, err)
	}
	_, err = cfgWriter.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to write config, update your config to %d manually. Error: %w", latestVer, err)
	}

	return db, nil
}

func CurrentSchema() []string {
	return []string{
		// DDL
		`CREATE TABLE user (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT
		)`,
		`CREATE TABLE user_platform (
			user_id TEXT NOT NULL,
			platform_id INTEGER NOT NULL,
			bot_is_joined BOOL NOT NULL DEFAULT false,
			PRIMARY KEY (user_id, platform_id),
			FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
			FOREIGN KEY (platform_id) REFERENCES platform(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE platform (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		)`,
		`CREATE TABLE command (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		)`,
		`CREATE TABLE user_platform_command (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			command_id INTEGER NOT NULL,
			platform_id INTEGER NOT NULL,
			FOREIGN KEY (user_id, platform_id, command_id) REFERENCES user_platform(user_id, platform_id, command_id) ON DELETE CASCADE
		)`,

		// DML
		`INSERT INTO platform (name) VALUES ('twitch')`,
	}
}

func InsertCommands(tx *sql.Tx, commands []Command) error {
	var (
		id  int
		err error
	)
	// check if commands were already added(expected to return ErrNoRows)
	err = tx.QueryRow("SELECT id FROM command LIMIT 1").Scan(&id)
	if err != sql.ErrNoRows {
		return fmt.Errorf("failed to get command: %w", err)
	}

	for _, command := range commands {
		_, err = tx.Exec("INSERT INTO command (name) VALUES (?)", command.Name)
		if err != nil {
			return fmt.Errorf("failed to insert command: %w", err)
		}
	}

	return nil
}

// Run migrations in the database.
// If the migration succeeds, the version in DBConfig is updated to the current version
// and should be saved in the config file.
func RunMigrations(tx *sql.Tx, config *Config, migrations *DBMigrations) error {
	// sort migrations by version
	sort.Sort(migrations)

	var err error

	// migrations to be applied sequentially until the currentVersion
	migrationsApplied := 0
	currentVersion := config.DBConfig.Version
	for i := currentVersion; i < len(migrations.Migrations); i++ {
		migration := migrations.Migrations[i]

		for _, stmt := range migration.Stmts {
			_, err = tx.Exec(stmt)
			if err != nil {
				return fmt.Errorf("failed to execute migration: %w", err)
			}
		}
		migrationsApplied++

		// version 1 creates the database from scratch so there's no need to run the other migrations,
		// and we can just update the version to the latest one.
		if migration.Version == 1 {
			currentVersion = migrations.Migrations[len(migrations.Migrations)-1].Version
			break
		}
	}

	if migrationsApplied == 0 {
		return nil
	}

	config.DBConfig.Version = migrations.Migrations[len(migrations.Migrations)-1].Version
	return nil
}
