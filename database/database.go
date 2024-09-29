package database

import (
	"database/sql"
	"fmt"
	"io"
	"monkebot/config"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/rs/zerolog/log"
)

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

	if driver == "sqlite3" {
		pragmas := []string{
			"PRAGMA main.page_size=8192;",
			"PRAGMA main.cache_size=15000;",
			"PRAGMA main.synchronous=NORMAL;",
			"PRAGMA main.journal_mode=WAL;",
			"PRAGMA main.temp_store=MEMORY;",
		}

		for _, p := range pragmas {
			_, err = db.Exec(p)
			if err != nil {
				return nil, fmt.Errorf("failed to set pragma: %w", err)
			}
		}
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

	// prepared statement to check if user command already exists
	var userCommandExistsStmt *sql.Stmt
	userCommandExistsStmt, err = tx.Prepare("SELECT id FROM user_command WHERE user_id = ? AND command_id = ?")
	commandExists := func(userID string, commandID int) bool {
		var id int
		err = userCommandExistsStmt.QueryRow(userID, commandID).Scan(&id)
		if err != nil && err != sql.ErrNoRows {
			return false
		}
		return err == nil
	}

	var userCommandInsertStmt *sql.Stmt
	userCommandInsertStmt, err = tx.Prepare("INSERT INTO user_command (user_id, command_id) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare user command insert: %w", err)
	}
	for _, commandID := range commandIDs {
		if commandExists(userID, commandID) {
			return fmt.Errorf("user %s already has command %d", userID, commandID)
		}
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
			log.Warn().Err(err).Str("name", user.Name).Msg("skipping insertion for user")
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

func UpdateUserPermission(tx *sql.Tx, username string, permissionName string) error {
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

	var res sql.Result
	res, err = tx.Exec("UPDATE user SET permission_id = ? WHERE name = ?", newPermID, username)
	if err != nil {
		return fmt.Errorf("failed to update user %s: %w", username, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected != 1 {
		return fmt.Errorf("invalid number of affected rows %d trying to update user %s's permission to %s", rowsAffected, username, permissionName)
	}

	return nil
}

func UpdateIsBotJoined(tx *sql.Tx, joined bool, userIDs ...string) error {
	stmt, err := tx.Prepare("UPDATE user SET bot_is_joined = ? WHERE id = ? ")
	if err != nil {
		return fmt.Errorf("failed to update is_bot_joined for user %s: %w", userIDs, err)
	}

	var res sql.Result
	for _, userID := range userIDs {
		res, err = stmt.Exec(joined, userID)
		if err != nil {
			return fmt.Errorf("failed to update is_bot_joined for user %s: %w", userID, err)
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected != 1 {
			return fmt.Errorf("invalid number of affected rows %d trying to update is_bot_joined to %t for user %s", rowsAffected, joined, userID)
		}

	}

	return nil
}

func UpdateIsUserCommandEnabled(tx *sql.Tx, enabled bool, channelID string, commandName string) error {
	result, err := tx.Exec(`
			UPDATE user_command SET is_enabled = ?
			WHERE command_id = (
				SELECT id FROM command WHERE name = ?
			) AND user_id = ?`,
		enabled, commandName, channelID)
	if err != nil {
		return fmt.Errorf("failed to update is_user_command_enabled: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected != 1 {
		return fmt.Errorf("invalid number of affected rows %d trying to update command %s to %t in channel %s", rowsAffected, commandName, enabled, channelID)
	}

	return nil
}

func SelectUserExists(tx *sql.Tx, username string) (bool, error) {
	var exists bool
	err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM user WHERE name = ?)", username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to select user exists: %w", err)
	}
	return exists, nil
}

func SelectIsUserCommandEnabled(tx *sql.Tx, channelID string, commandName string) (bool, error) {
	var enabled bool
	err := tx.QueryRow(`
			SELECT is_enabled
			FROM user_command
			WHERE command_id = (
				SELECT id FROM command WHERE name = ?
			) AND user_id = ?`,
		commandName, channelID).Scan(&enabled)
	if err != nil {
		return false, fmt.Errorf("failed to select is_user_command_enabled: %w", err)
	}

	return enabled, nil
}

func SelectIsCommandOnCooldown(tx *sql.Tx, channelID string, commandName string, cooldown int) (bool, error) {
	var lastUsed int64
	err := tx.QueryRow(`
			SELECT uc.last_used
			FROM user_command uc
			INNER JOIN command c ON c.id = uc.command_id
			WHERE c.name = ? AND uc.user_id = ?`,
		commandName, channelID).Scan(&lastUsed)
	if err != nil {
		return false, fmt.Errorf("failed to select is_command_on_cooldown: %w", err)
	}

	cooldownDuration, err := time.ParseDuration(fmt.Sprintf("%ds", cooldown))
	if err != nil {
		return false, fmt.Errorf("failed to parse cooldown duration: %w", err)
	}
	lastUsedTime := time.Unix(lastUsed, 0)
	return time.Now().Before(lastUsedTime.Add(cooldownDuration)), nil
}

func UpdateUserCommandLastUsed(tx *sql.Tx, channelID string, commandName string) error {
	var (
		err error
		id  int
	)
	err = tx.QueryRow(`
		SELECT uc.id FROM user_command uc
		INNER JOIN command c ON c.id = uc.command_id
		WHERE c.name = ? AND uc.user_id = ?
		`, commandName, channelID).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to update command's %s cooldown: %w", commandName, err)
	}

	var result sql.Result
	result, err = tx.Exec("UPDATE user_command SET last_used = ? WHERE id = ?", time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("failed to update command's %s cooldown: %w", commandName, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected != 1 {
		return fmt.Errorf("invalid number of affected rows trying to update command's %s cooldown: %d", commandName, rowsAffected)
	}

	return nil
}
