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

func generateTestConfig() (string, error) {
	tempFile, err := os.CreateTemp(os.TempDir(), "monkebotTestJson")
	if err != nil {
		return "", err
	}
	template, err := ConfigTemplateJSON()
	if err != nil {
		return "", err
	}
	err = os.WriteFile(tempFile.Name(), template, 0644)
	if err != nil {
		return "", err
	}
	_, err = LoadConfigFromFile(tempFile.Name())
	if err != nil {
		return "", err
	}
	return tempFile.Name(), nil
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

	filename, err = generateTestConfig()
	if err != nil {
		t.Errorf("failed to generate test config: %v", err)
	}
	err = RunMigrations(db, filename, &migrations)
	if err != nil {
		t.Errorf("failed to run migrations with current schema: %v", err)
	}
}
