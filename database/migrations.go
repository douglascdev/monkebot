package database

import (
	"database/sql"
	"fmt"
	"monkebot/config"
	"sort"
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
		{Version: 2, Stmts: []string{
			"INSERT INTO command (name) VALUES ('butt')",
			`INSERT INTO user_command (user_id, command_id, is_enabled)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'butt'
				), true FROM user
			`,
		}},
		{Version: 3, Stmts: []string{
			"INSERT INTO command (name) VALUES ('help')",
			`INSERT INTO user_command (user_id, command_id, is_enabled)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'help'
				), true FROM user
			`,
		}},
		{Version: 4, Stmts: []string{
			"INSERT INTO command (name) VALUES ('explore')",
			`INSERT INTO user_command (user_id, command_id, is_enabled)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'explore'
				), true FROM user`,
			`CREATE TABLE rpg_item (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				description TEXT NOT NULL
			)`,
			`CREATE TABLE rpg_user_item (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				user_id TEXT NOT NULL,
				rpg_item_id INTEGER NOT NULL,
				amount INTEGER NOT NULL,

				FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
				FOREIGN KEY (rpg_item_id) REFERENCES rpg_item(id) ON DELETE CASCADE
			)`,
			`INSERT INTO rpg_item (name, description) VALUES ('buttinho', 'The most widely used currency in the seven seas.')`,
		}},
		{Version: 5, Stmts: []string{
			"INSERT INTO command (name) VALUES ('enable'), ('disable')",
			`INSERT INTO user_command (user_id, command_id, is_enabled)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'enable'
				), true FROM user`,
			`INSERT INTO user_command (user_id, command_id, is_enabled)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'disable'
				), true FROM user`,
		}},
		{Version: 6, Stmts: []string{
			"ALTER TABLE user_command ADD last_used INTEGER NOT NULL DEFAULT 1726849749",
		}},
		{Version: 7, Stmts: []string{
			`CREATE INDEX idx_user_command ON user_command(user_id, command_id)`,
			`CREATE TABLE user_command_cooldown (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				command_id INTEGER NOT NULL,
				last_used INTEGER NOT NULL DEFAULT 1726849749,
				FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
				FOREIGN KEY (command_id) REFERENCES command(id) ON DELETE CASCADE
			)`,
			`CREATE INDEX idx_user_command_cooldown ON user_command_cooldown(user_id, command_id, last_used)`,
			`
			INSERT INTO user_command_cooldown (user_id, command_id)
			SELECT u.id, c.id
			FROM user u
			CROSS JOIN command c
			`,
		}},
	},
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
