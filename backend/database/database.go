package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(dbPath string) {
	var err error
	DB, err = sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	createTables()
	log.Println("Database initialized successfully")
}

func createTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			provider_type TEXT NOT NULL,
			api_key TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS operation_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			action TEXT NOT NULL,
			resource TEXT NOT NULL,
			resource_id TEXT NOT NULL,
			details TEXT,
			ip_address TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_user_id ON operation_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_created_at ON operation_logs(created_at DESC)`,
		`CREATE TABLE IF NOT EXISTS domain_cache (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			account_id INTEGER NOT NULL,
			domain_id TEXT NOT NULL,
			domain_name TEXT NOT NULL,
			renewal_date TEXT DEFAULT '',
			renewal_url TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE(user_id, account_id, domain_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_domain_cache_user_id ON domain_cache(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_domain_cache_domain_name ON domain_cache(domain_name)`,

		// Notification settings table
		`CREATE TABLE IF NOT EXISTS notification_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			domain_id TEXT NOT NULL,
			account_id INTEGER NOT NULL,
			days_before INTEGER NOT NULL DEFAULT 30,
			enabled INTEGER NOT NULL DEFAULT 1,
			last_notified_at DATETIME,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			UNIQUE(user_id, account_id, domain_id),
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_notification_settings_user_id ON notification_settings(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_notification_settings_enabled ON notification_settings(enabled)`,

		// Email configuration table
		`CREATE TABLE IF NOT EXISTS email_config (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL UNIQUE,
			smtp_host TEXT NOT NULL,
			smtp_port INTEGER NOT NULL,
			smtp_username TEXT NOT NULL,
			smtp_password TEXT NOT NULL,
			from_email TEXT NOT NULL,
			from_name TEXT,
			to_email TEXT NOT NULL,
			enabled INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_email_config_user_id ON email_config(user_id)`,

		// Scheduler logs table
		`CREATE TABLE IF NOT EXISTS scheduler_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_name TEXT NOT NULL,
			status TEXT NOT NULL,
			message TEXT,
			details TEXT,
			started_at DATETIME NOT NULL,
			completed_at DATETIME,
			duration_ms INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_scheduler_logs_task_name ON scheduler_logs(task_name)`,
		`CREATE INDEX IF NOT EXISTS idx_scheduler_logs_created_at ON scheduler_logs(created_at DESC)`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			log.Fatalf("Failed to create table: %v", err)
		}
	}
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
