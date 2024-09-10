package monkebot

import (
	"bytes"
	"database/sql"
	"io"
	"testing"
)

func generateTestDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file:data.db?mode=memory")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func generateTestConfig() (*Config, error) {
	template, err := ConfigTemplateJSON()
	if err != nil {
		return nil, err
	}
	var cfg *Config
	cfg, err = LoadConfig(template)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func TestGenerateTestDB(t *testing.T) {
	db, err := generateTestDB()
	if err != nil {
		t.Errorf("failed to init test database: %v", err)
	}
	defer db.Close()
}

func TestInitDB(t *testing.T) {
	cfg, err := generateTestConfig()
	if err != nil {
		t.Errorf("failed to generate test config: %v", err)
	}
	cfg.DBConfig.Version = 0

	var (
		reader = new(bytes.Buffer)
		writer = new(bytes.Buffer)
		data   []byte
	)

	data, err = MarshalConfig(cfg)
	if err != nil {
		t.Errorf("failed to marshal test config: %v", err)
	}
	reader.Write(data)

	db, err := InitDB("sqlite3", "file:data.db?mode=memory", reader, writer)
	if err != nil {
		t.Errorf("failed to run InitDB: %v", err)
	}
	defer db.Close()

	data, err = io.ReadAll(writer)
	if err != nil {
		t.Errorf("failed to read written config: %v", err)
	}

	cfg, err = LoadConfig(data)
	if err != nil {
		t.Errorf("failed to load written config: %v", err)
	}
	if cfg.DBConfig.Version != 1 {
		t.Errorf("migration failed to update database version, expected 1, got %d", cfg.DBConfig.Version)
	}
}

func TestRunMigrationsCurrentSchema(t *testing.T) {
	db, _ := generateTestDB()

	migrations := DBMigrations{
		Migrations: []DBMigration{
			{Version: 1, Stmts: CurrentSchema()},
		},
	}

	var (
		cfg *Config
		err error
	)

	cfg, err = generateTestConfig()
	if err != nil {
		t.Errorf("failed to generate test config: %v", err)
	}
	err = RunMigrations(db, cfg, &migrations)
	if err != nil {
		t.Errorf("failed to run migrations with current schema: %v", err)
	}
	res := db.QueryRow("SELECT id, name FROM platform")

	var (
		id       int
		platform string
	)
	err = res.Scan(&id, &platform)
	if err != nil {
		t.Errorf("failed to scan platform value: %v", err)
	}
	if platform != "twitch" {
		t.Errorf("unexpected platform value: %s", platform)
	}
}

func TestRunMigrationsCurrentSchemaAndNewMigrations(t *testing.T) {
	db, _ := generateTestDB()

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
		cfg *Config
		err error
	)

	cfg, err = generateTestConfig()
	if err != nil {
		t.Errorf("failed to generate test config: %v", err)
	}
	err = RunMigrations(db, cfg, &migrations)
	if err != nil {
		t.Errorf("failed to run migrations with current schema: %v", err)
	}

	if cfg.DBConfig.Version != 3 {
		t.Errorf("expected version 3, got %d", cfg.DBConfig.Version)
	}
}

func TestRunMigrationsNewMigrations(t *testing.T) {
	db, _ := generateTestDB()

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
		cfg *Config
		err error
	)

	cfg, err = generateTestConfig()
	if err != nil {
		t.Errorf("failed to generate test config: %v", err)
	}
	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT NOT NULL)")
	if err != nil {
		t.Errorf("failed to create test table: %v", err)
	}

	cfg.DBConfig.Version = 2

	err = RunMigrations(db, cfg, &migrations)
	if err != nil {
		t.Errorf("failed to run migrations: %v", err)
	}

	if cfg.DBConfig.Version != 3 {
		t.Errorf("expected version 3, got %d", cfg.DBConfig.Version)
	}

	res := db.QueryRow("SELECT id, name FROM test")
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

func TestInsertCommands(t *testing.T) {
	db, _ := generateTestDB()

	migrations := DBMigrations{
		Migrations: []DBMigration{
			{Version: 1, Stmts: CurrentSchema()},
		},
	}

	var (
		cfg *Config
		err error
	)

	cfg, err = generateTestConfig()
	if err != nil {
		t.Errorf("failed to generate test config: %v", err)
	}

	err = RunMigrations(db, cfg, &migrations)
	if err != nil {
		t.Errorf("failed to run migrations: %v", err)
	}

	err = InsertCommands(db, []Command{
		{Name: "test"},
	})
	if err != nil {
		t.Errorf("failed to insert commands: %v", err)
	}

	res := db.QueryRow("SELECT name FROM command")
	var name string
	err = res.Scan(&name)
	if err != nil {
		t.Errorf("failed to scan name value: %v", err)
	}
	if name != "test" {
		t.Errorf("unexpected name value: %s", name)
	}
}
