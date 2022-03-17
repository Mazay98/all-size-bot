package config

import (
	"os"
	"strconv"
)

type AppConfig struct {
	Telegram TelegramConfig
	Postgres Postgres
}
type TelegramConfig struct {
	Token    string
	BotDebug bool
}
type Postgres struct {
	DbConnectionString string
}

func New() *AppConfig {
	return &AppConfig{
		Telegram: TelegramConfig{
			Token:    getEnv("TELEGRAM_API", ""),
			BotDebug: getEnvAsBool("BOT_DEBUG", true),
		},
		Postgres: Postgres{
			DbConnectionString: getEnv("DB_DATABASE_CONNECTION_STRING", ""),
		},
	}
}

// Helper to read an environment variable into a string or return default value.
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

// Helper to read an environment variable into a bool or return default value.
func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}
