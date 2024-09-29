package database

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
			last_used INTEGER NOT NULL DEFAULT 1726849749,
			FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
			FOREIGN KEY (command_id) REFERENCES command(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX idx_is_enabled ON user_command(is_enabled)`,
		`CREATE INDEX idx_user_command ON user_command(user_id, command_id)`,
		`CREATE TABLE user_command_data (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			command_id INTEGER NOT NULL,
			last_used INTEGER NOT NULL DEFAULT 1726849749,
			FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
			FOREIGN KEY (command_id) REFERENCES command(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX idx_user_command_data ON user_command_data(user_id, command_id, last_used)`,
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

		// DML
		`INSERT INTO permission (name) VALUES ('user')`,
		`INSERT INTO permission (name, is_ignored) VALUES ('banned', true)`,
		`INSERT INTO permission (name, is_bot_admin) VALUES ('admin', true)`,

		`INSERT INTO rpg_item (name, description) VALUES ('buttinho', 'The most widely used currency in the seven seas.')`,
	}
}
