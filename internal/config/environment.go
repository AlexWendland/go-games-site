package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Production     bool
	Port           int
	BaseURL        string
	AllowedOrigins []string
	DatabaseFile   string
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, ok := os.LookupEnv(key); ok {
		num, err := strconv.Atoi(value)
		if err != nil {
			slog.Error("failed to parse env var as int", "key", key, "value", value)
			os.Exit(1)
		}
		return num
	}
	return defaultValue

}

func getEnvList(key string, defaultValue []string) []string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.Split(value, ",")
	}
	return defaultValue
}

func printConfig(config Config) {
	slog.Info("config loaded",
		"production", config.Production,
		"port", config.Port,
		"base_url", config.BaseURL,
		"allowed_origins", config.AllowedOrigins,
		"database_file", config.DatabaseFile,
	)
}

func Load() Config {
	config := Config{
		Production:     getEnv("APP_ENV", "development") == "production",
		Port:           getEnvInt("PORT", 8080),
		BaseURL:        getEnv("BASE_URL", "http://localhost:8080"),
		AllowedOrigins: getEnvList("ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
		DatabaseFile:   getEnv("DATABASE_FILE", "games.db"),
	}
	printConfig(config)
	return config
}
