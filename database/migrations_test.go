package database

import (
	"monkebot/config"
	"testing"
)

func TestRunMigrationsCurrentSchema(t *testing.T) {
	migrations := DBMigrations{
		Migrations: []DBMigration{
			{Version: 1, Stmts: CurrentSchema()},
		},
	}

	var (
		cfg *config.Config
		err error
	)

	cfg, err = generateTestConfig()
	if err != nil {
		t.Errorf("failed to generate test config: %v", err)
	}
	tx, err := testDB.Begin()
	defer tx.Rollback()
	if err != nil {
		t.Errorf("failed to begin transaction: %v", err)
	}
	err = RunMigrations(tx, cfg, &migrations)
	if err != nil {
		t.Errorf("failed to run migrations with current schema: %v", err)
	}
	res := tx.QueryRow("SELECT id FROM permission WHERE name = 'user'")

	var id int
	err = res.Scan(&id)
	if err != nil {
		t.Errorf("failed to scan permission value: %v", err)
	}
}

func TestRunMigrationsCurrentSchemaAndNewMigrations(t *testing.T) {
	migrations := DBMigrations{
		Migrations: []DBMigration{
			{Version: 1, Stmts: CurrentSchema()},
			{Version: 2, Stmts: []string{
				"CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT NOT NULL)",
			}},
			{Version: 3, Stmts: []string{
				"INSERT INTO test (name) VALUES ('test')",
			}},
		},
	}

	var (
		cfg *config.Config
		err error
	)

	cfg, err = generateTestConfig()
	if err != nil {
		t.Errorf("failed to generate test config: %v", err)
	}

	tx, err := testDB.Begin()
	defer tx.Rollback()
	if err != nil {
		t.Errorf("failed to begin transaction: %v", err)
	}
	err = RunMigrations(tx, cfg, &migrations)
	if err != nil {
		t.Errorf("failed to run migrations with current schema: %v", err)
	}

	if cfg.DBConfig.Version != 3 {
		t.Errorf("expected version 3, got %d", cfg.DBConfig.Version)
	}
}

func TestRunMigrationsNewMigrations(t *testing.T) {
	migrations := DBMigrations{
		Migrations: []DBMigration{
			{Version: 1, Stmts: CurrentSchema()},
			{Version: 2, Stmts: []string{
				"CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT NOT NULL)",
			}},
			{Version: 3, Stmts: []string{
				"INSERT INTO test (name) VALUES ('test')",
			}},
		},
	}

	var (
		cfg *config.Config
		err error
	)

	tx, err := testDB.Begin()
	defer tx.Rollback()
	if err != nil {
		t.Errorf("failed to begin transaction: %v", err)
	}

	cfg, err = generateTestConfig()
	if err != nil {
		t.Errorf("failed to generate test config: %v", err)
	}
	_, err = tx.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT NOT NULL)")
	if err != nil {
		t.Errorf("failed to create test table: %v", err)
	}

	cfg.DBConfig.Version = 2

	err = RunMigrations(tx, cfg, &migrations)
	if err != nil {
		t.Errorf("failed to run migrations: %v", err)
	}

	if cfg.DBConfig.Version != 3 {
		t.Errorf("expected version 3, got %d", cfg.DBConfig.Version)
	}

	res := tx.QueryRow("SELECT id, name FROM test")
	var (
		id   int
		name string
	)
	err = res.Scan(&id, &name)
	if err != nil {
		t.Errorf("failed to scan name value: %v", err)
	}
	if name != "test" {
		t.Errorf("unexpected name value: %s", name)
	}
}
