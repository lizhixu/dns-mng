package database

import (
	"database/sql"
	"log"
	"strings"

	"github.com/tursodatabase/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

// dbDriver 记录当前使用的驱动名 ("sqlite" 或 "libsql")，供 VACUUM 等仅本地可用的操作判断。
var dbDriver string

// IsLibSQL 返回当前是否使用 libSQL 驱动 (Turso/远程)。
func IsLibSQL() bool { return dbDriver == "libsql" }

// InitWithConfig 根据配置选择驱动并初始化数据库连接。
// dbType: "sqlite" (本地文件, 默认) 或 "libsql" (Turso/远程 libSQL)。
func InitWithConfig(dbType, dbPath, dbURL, dbAuthToken string) {
	var err error

	switch strings.ToLower(dbType) {
	case "", "sqlite":
		dbDriver = "sqlite"
		DB, err = sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
		if err != nil {
			log.Fatalf("Failed to open sqlite database: %v", err)
		}
		if err = DB.Ping(); err != nil {
			log.Fatalf("Failed to ping sqlite database: %v", err)
		}
		// SQLite is embedded and this application performs async API log writes while
		// pages are reading log data. Keeping a single connection avoids intermittent
		// "database is locked" failures from competing pooled connections.
		DB.SetMaxOpenConns(1)
		DB.SetMaxIdleConns(1)
	case "libsql":
		dbDriver = "libsql"
		var opts []libsql.Option
		if dbAuthToken != "" {
			opts = append(opts, libsql.WithAuthToken(dbAuthToken))
		}
		connector, err := libsql.NewConnector(dbURL, opts...)
		if err != nil {
			log.Fatalf("Failed to create libsql connector: %v", err)
		}
		DB = sql.OpenDB(connector)
		if err = DB.Ping(); err != nil {
			log.Fatalf("Failed to ping libsql database: %v", err)
		}
		// 远程/嵌入式 libSQL 不存在本地文件锁竞争，放开连接池以提升并发吞吐。
		DB.SetMaxOpenConns(10)
		DB.SetMaxIdleConns(5)
		DB.SetConnMaxIdleTime(0)
	default:
		log.Fatalf("Unsupported DB_TYPE %q: expected \"sqlite\" or \"libsql\"", dbType)
	}

	createTables()

	// Run migration from old operation_logs to new api_call_logs
	if err := MigrateOperationLogsToAPILogs(); err != nil {
		log.Printf("Warning: Migration failed: %v", err)
	}

	log.Printf("Database initialized successfully (driver=%s)", dbDriver)
}

// Init 保留旧签名以兼容外部调用 (按本地 sqlite 路径初始化)。
// Deprecated: 请改用 InitWithConfig。
func Init(dbPath string) {
	InitWithConfig("sqlite", dbPath, "", "")
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
		// New API call logs table - complete recording without truncation
		`CREATE TABLE IF NOT EXISTS api_call_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			method TEXT NOT NULL,
			path TEXT NOT NULL,
			query TEXT,
			request_headers TEXT,
			request_body TEXT,
			status_code INTEGER NOT NULL,
			response_body TEXT,
			ip_address TEXT,
			user_agent TEXT,
			duration_ms INTEGER,
			error_message TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_api_logs_user_id ON api_call_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_api_logs_created_at ON api_call_logs(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_api_logs_user_created ON api_call_logs(user_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_api_logs_method ON api_call_logs(method)`,
		`CREATE INDEX IF NOT EXISTS idx_api_logs_path ON api_call_logs(path)`,
		`CREATE INDEX IF NOT EXISTS idx_api_logs_status_code ON api_call_logs(status_code)`,
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
		// 为兼容旧版本，添加列（如果不存在）
		`ALTER TABLE domain_cache ADD COLUMN deleted_at DATETIME`,
		`ALTER TABLE domain_cache ADD COLUMN last_sync_at DATETIME`,
		`ALTER TABLE domain_cache ADD COLUMN provider_updated_on DATETIME`,
		`CREATE INDEX IF NOT EXISTS idx_domain_cache_user_id ON domain_cache(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_domain_cache_domain_name ON domain_cache(domain_name)`,
		`CREATE INDEX IF NOT EXISTS idx_domain_cache_deleted_at ON domain_cache(deleted_at)`,

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
		// 为兼容旧版本，添加邮件语言列（如果不存在）
		`ALTER TABLE email_config ADD COLUMN language TEXT DEFAULT ''`,

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

		// DDNS tokens table (one token per user, system-wide)
		`CREATE TABLE IF NOT EXISTS ddns_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL UNIQUE,
			token TEXT NOT NULL UNIQUE,
			enabled INTEGER DEFAULT 1,
			last_used_at DATETIME,
			last_ip TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_ddns_tokens_token ON ddns_tokens(token)`,

		// Login logs table
		`CREATE TABLE IF NOT EXISTS login_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL DEFAULT 0,
			username TEXT NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			device TEXT,
			status TEXT NOT NULL,
			message TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_login_logs_user_id ON login_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_login_logs_created_at ON login_logs(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_login_logs_username ON login_logs(username)`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			// ALTER TABLE ADD COLUMN 在列已存在时会报错，忽略此错误
			if strings.Contains(q, "ALTER TABLE") && strings.Contains(err.Error(), "duplicate column name") {
				continue
			}
			log.Fatalf("Failed to create table: %v", err)
		}
	}
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
