package monkebot

import (
	"database/sql"
	"os"
	"testing"
)

func generateTestDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file:data.db?mode=memory")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func generateTestConfig() (*Config, string, error) {
	tempFile, err := os.CreateTemp(os.TempDir(), "monkebotTestJson")
	if err != nil {
		return nil, "", err
	}
	template, err := ConfigTemplateJSON()
	if err != nil {
		return nil, "", err
	}
	err = os.WriteFile(tempFile.Name(), template, 0644)
	if err != nil {
		return nil, "", err
	}
	var cfg *Config
	cfg, err = LoadConfigFromFile(tempFile.Name())
	if err != nil {
		return nil, "", err
	}
	return cfg, tempFile.Name(), nil
}

func TestInitDB(t *testing.T) {
	db, err := generateTestDB()
	if err != nil {
		t.Errorf("failed to init test database: %v", err)
	}
	defer db.Close()
}

func TestRunMigrationsCurrentSchema(t *testing.T) {
	db, _ := generateTestDB()

	migrations := DBMigrations{
		Migrations: []DBMigration{
			{Version: 1, Stmts: CurrentSchema()},
		},
	}

	var (
		err      error
		filename string
	)

	_, filename, err = generateTestConfig()
	if err != nil {
		t.Errorf("failed to generate test config: %v", err)
	}
	err = RunMigrations(db, filename, &migrations)
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
		err      error
		filename string
	)

	_, filename, err = generateTestConfig()
	if err != nil {
		t.Errorf("failed to generate test config: %v", err)
	}
	err = RunMigrations(db, filename, &migrations)
	if err != nil {
		t.Errorf("failed to run migrations with current schema: %v", err)
	}

	var cfg *Config
	cfg, err = LoadConfigFromFile(filename)
	if err != nil {
		t.Errorf("failed to load config: %v", err)
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
		cfg      *Config
		err      error
		filename string
	)

	cfg, filename, err = generateTestConfig()
	if err != nil {
		t.Errorf("failed to generate test config: %v", err)
	}
	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT NOT NULL)")
	if err != nil {
		t.Errorf("failed to create test table: %v", err)
	}

	cfg.DBConfig.Version = 2
	err = SaveConfigToFile(cfg, filename)
	if err != nil {
		t.Errorf("failed to save config: %v", err)
	}

	err = RunMigrations(db, filename, &migrations)
	if err != nil {
		t.Errorf("failed to run migrations: %v", err)
	}

	cfg, err = LoadConfigFromFile(filename)
	if err != nil {
		t.Errorf("failed to load config: %v", err)
	}
	if cfg.DBConfig.Version != 3 {
		t.Errorf("expected version 3, got %d", cfg.DBConfig.Version)
	}

	res := db.QueryRow("SELECT * FROM test")
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
		err      error
		filename string
	)

	_, filename, err = generateTestConfig()
	if err != nil {
		t.Errorf("failed to generate test config: %v", err)
	}

	err = RunMigrations(db, filename, &migrations)
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
