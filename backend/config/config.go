package config

import (
	"os"
)

type Config struct {
	ServerPort string
	DBPath     string
	JWTSecret  string
}

func Load() *Config {
	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		DBPath:     getEnv("DB_PATH", "dns-mng.db"),
		JWTSecret:  getEnv("JWT_SECRET", "dns-mng-secret-key-change-in-production"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
