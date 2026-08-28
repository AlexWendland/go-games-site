package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/AlexWendland/go-games-site/internal/api"
	"github.com/AlexWendland/go-games-site/internal/db"
	"github.com/AlexWendland/go-games-site/internal/db/migrations"
	"github.com/AlexWendland/go-games-site/internal/web"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	server := http.NewServeMux()

	// Set up database
	// TODO: Migrate to using an environment variable for db file name.
	dbConnection, err := db.Open("games.db", migrations.FS)
	if err != nil {
		logger.Error("can not open database connection", slog.Any("error", err))
		return
	}
	defer dbConnection.Close()

	// Add handlers here
	server.Handle("/", web.WebServerHandler())
	server.Handle("/api/", http.StripPrefix("/api", api.ApiHandler(dbConnection)))

	logger.Info("hosting server on http://localhost:8080/")
	if err := http.ListenAndServe(":8080", server); err != nil {
		logger.Error("server shut down unexpectedly", slog.Any("error", err))
	}
}
