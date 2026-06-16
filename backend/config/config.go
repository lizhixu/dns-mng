package config

import (
	"os"
)

type Config struct {
	ServerPort  string
	DBType      string // "sqlite" (默认, 本地文件) 或 "libsql" (Turso/远程 libSQL)
	DBPath      string // 仅 DBType=sqlite 时使用
	DBURL       string // DBType=libsql 时使用, 如 libsql://xxx.turso.io 或 file:./local.db
	DBAuthToken string // DBType=libsql 时使用, Turso 访问令牌 (本地文件可留空)
	JWTSecret   string
}

func Load() *Config {
	return &Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		DBType:      getEnv("DB_TYPE", "sqlite"),
		DBPath:      getEnv("DB_PATH", "dns-mng.db"),
		DBURL:       getEnv("DB_URL", ""),
		DBAuthToken: getEnv("DB_AUTH_TOKEN", ""),
		JWTSecret:   getEnv("JWT_SECRET", "dns-mng-secret-key-change-in-production"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
