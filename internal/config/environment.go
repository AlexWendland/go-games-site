package config

import (
	"log/slog"

	"github.com/kelseyhightower/envconfig"
)

// See below for syntax:
// https://pkg.go.dev/github.com/kelseyhightower/envconfig#section-readme
type Config struct {
	Production     bool     `default:"false"`
	Port           int      `default:"8080"`
	BaseURL        string   `default:"http://localhost:8080" split_words:"true"`
	AllowedOrigins []string `default:"http://localhost:5173" split_words:"true"`
	DatabaseFile   string   `default:"games.db" split_words:"true"`
}

func PrintConfig(config *Config) {
	slog.Info("current config",
		"production", config.Production,
		"port", config.Port,
		"base_url", config.BaseURL,
		"allowed_origins", config.AllowedOrigins,
		"database_file", config.DatabaseFile,
	)
}

func Load() (*Config, error) {
	var this Config
	if err := envconfig.Process("games", &this); err != nil {
		return nil, err
	}
	return &this, nil
}
