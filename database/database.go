package database

import (
	"database/sql"
	"fmt"
	"io"
	"monkebot/config"
	"sort"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/rs/zerolog/log"
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

var Migrations = DBMigrations{
	Migrations: []DBMigration{
		{Version: 1, Stmts: CurrentSchema()},
	},
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

	var cfg *config.Config
	var data []byte
	data, err = io.ReadAll(cfgReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg, err = config.LoadConfig(data)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	err = RunMigrations(tx, cfg, &Migrations)
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	latestVer := Migrations.Migrations[len(Migrations.Migrations)-1].Version
	data, err = config.MarshalConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config, update your config to %d manually. Error: %w", latestVer, err)
	}
	_, err = cfgWriter.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to write config, update your config to %d manually. Error: %w", latestVer, err)
	}

	return db, nil
}

func IsSchemaUpToDate(cfg *config.Config) bool {
	return cfg.DBConfig.Version == len(Migrations.Migrations)
}

func CurrentSchema() []string {
	return []string{
		// DDL
		`CREATE TABLE user (
			id TEXT NOT NULL PRIMARY KEY,
			name TEXT NOT NULL,
			permission_id INTEGER NOT NULL,
			bot_is_joined BOOL NOT NULL DEFAULT false,
			FOREIGN KEY (permission_id) REFERENCES permission(id)
		)`,
		`CREATE TABLE permission (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			is_ignored BOOL NOT NULL DEFAULT false,
			is_bot_admin BOOL NOT NULL DEFAULT false
		)`,
		`CREATE TABLE command (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		)`,
		`CREATE INDEX idx_name ON command(name)`,
		`CREATE TABLE user_command (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			command_id INTEGER NOT NULL,
			is_enabled BOOL NOT NULL DEFAULT true,
			FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
			FOREIGN KEY (command_id) REFERENCES command(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX idx_is_enabled ON user_command(is_enabled)`,

		// DML
		`INSERT INTO permission (name) VALUES ('user')`,
		`INSERT INTO permission (name, is_ignored) VALUES ('banned', true)`,
		`INSERT INTO permission (name, is_bot_admin) VALUES ('admin', true)`,
	}
}

func SelectIsUserIgnored(tx *sql.Tx, userID string) (bool, error) {
	var (
		err       error
		isIgnored bool
	)

	err = tx.QueryRow(`
		SELECT p.is_ignored FROM permission p
		INNER JOIN user u ON u.permission_id = p.id
		WHERE u.id = ?
	`, userID).Scan(&isIgnored)
	if err != nil {
		return false, err
	}

	return isIgnored, nil
}

func SelectJoinedChannels(tx *sql.Tx) ([]string, error) {
	var (
		err      error
		channels []string
	)
	rows, err := tx.Query("SELECT name FROM user WHERE bot_is_joined")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var channel string
		err = rows.Scan(&channel)
		if err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	return channels, nil
}

func SelectIsUserAdmin(tx *sql.Tx, userID string) (bool, error) {
	var (
		err     error
		isAdmin bool
	)
	err = tx.QueryRow(`
		SELECT p.is_bot_admin FROM permission p
		INNER JOIN user u ON u.permission_id = p.id
		WHERE u.id = ?
	`, userID).Scan(&isAdmin)
	if err != nil {
		return false, err
	}
	return isAdmin, nil
}

func InsertCommands(tx *sql.Tx, commandNames ...string) error {
	var (
		id  int
		err error
	)
	// check if commands were already added(expected to return ErrNoRows)
	err = tx.QueryRow("SELECT id FROM command LIMIT 1").Scan(&id)
	if err != sql.ErrNoRows {
		return fmt.Errorf("failed to get command: %w", err)
	}

	var preparedInsert *sql.Stmt
	preparedInsert, err = tx.Prepare("INSERT INTO command (name) VALUES (?)")
	if err != nil {
		return fmt.Errorf("failed to create prepared insert for commands: %w", err)
	}
	for _, command := range commandNames {
		_, err = preparedInsert.Exec(command)
		if err != nil {
			return fmt.Errorf("failed to insert command: %w", err)
		}
	}

	return nil
}

// inserts the current list of commands for a user, so that admins have channel-level control over commands
func InsertUserCommands(tx *sql.Tx, userID string, commandNames ...string) error {
	rows, err := tx.Query("SELECT id FROM command")
	if err != nil {
		return fmt.Errorf("failed to get command ids: %w", err)
	}
	defer rows.Close()

	commandIDs := make([]int, len(commandNames))
	for i := 0; rows.Next(); i++ {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			return fmt.Errorf("failed to scan command ids: %w", err)
		}
		commandIDs[i] = id
	}

	err = rows.Err()
	if err != nil {
		return fmt.Errorf("failed to get command ids: %w", err)
	}

	var userCommandInsertStmt *sql.Stmt
	userCommandInsertStmt, err = tx.Prepare("INSERT INTO user_command (user_id, command_id) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare user command insert: %w", err)
	}
	for _, commandID := range commandIDs {
		_, err = userCommandInsertStmt.Exec(userID, commandID)
		if err != nil {
			return fmt.Errorf("failed to insert user command: %w", err)
		}
	}

	return nil
}

// Users that already exist will be ignored.
// All PlatformUsers must belong to the same platform.
func InsertUsers(tx *sql.Tx, joinBot bool, users ...struct{ ID, Name string }) error {
	var (
		row *sql.Row
		err error
	)

	// find user permission id
	var userPermissionID int
	row = tx.QueryRow("SELECT id FROM permission WHERE name = ?", "user")
	err = row.Scan(&userPermissionID)
	if err != nil {
		return fmt.Errorf("failed to find user permission: %w", err)
	}

	// prepare user insert
	var userInsertStmt *sql.Stmt
	userInsertStmt, err = tx.Prepare("INSERT INTO user (id, permission_id, name, bot_is_joined) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare user insert: %w", err)
	}

	// insert users
	var (
		result sql.Result
		userID int64
	)
	for _, user := range users {
		result, err = userInsertStmt.Exec(user.ID, userPermissionID, user.Name, joinBot)
		if err != nil {
			log.Err(err).Str("name", user.Name).Msg("skipping insertion for user")
			continue
		}
		userID, err = result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get inserted user's id")
		}
		log.Info().Int64("user_id", userID).Str("name", user.Name).Msg("inserted new user")
	}
	return nil
}

func UpdateUserPermission(tx *sql.Tx, userID string, permissionName string) error {
	var (
		err       error
		newPermID int64
	)
	err = tx.QueryRow(`
		SELECT id FROM permission p WHERE p.name = ?
	`, permissionName).Scan(&newPermID)
	if err != nil {
		return fmt.Errorf("failed to find id for permission %s: %w", permissionName, err)
	}

	_, err = tx.Exec("UPDATE user SET permission_id = ? WHERE id = ?", newPermID, userID)
	if err != nil {
		return fmt.Errorf("failed to update user %s: %w", userID, err)
	}

	return nil
}

func UpdateIsBotJoined(tx *sql.Tx, joined bool, userIDs ...string) error {
	stmt, err := tx.Prepare("UPDATE user SET bot_is_joined = ? WHERE id = ? ")
	if err != nil {
		return fmt.Errorf("failed to update is_bot_joined for user %s: %w", userIDs, err)
	}

	for _, userID := range userIDs {
		_, err = stmt.Exec(joined, userID)
		if err != nil {
			return fmt.Errorf("failed to update is_bot_joined for user %s: %w", userID, err)
		}
	}

	return nil
}

// Run migrations in the database.
// If the migration succeeds, the version in DBConfig is updated to the current version
// and should be saved in the config file.
func RunMigrations(tx *sql.Tx, config *config.Config, migrations *DBMigrations) error {
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
