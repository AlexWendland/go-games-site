package main

import (
	"net/http"
	"log/slog"
	"os"

	"github.com/AlexWendland/go-games-site/internal/web"
	"github.com/AlexWendland/go-games-site/internal/api"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	server := http.NewServeMux()

	// Add handlers here
	server.Handle("/", web.WebServerHandler())
	server.Handle("/api/", http.StripPrefix("/api", api.ApiHandler()))

	logger.Info("hosting server on http://localhost:8080/")
	if err := http.ListenAndServe(":8080", server); err != nil {
		logger.Error("server shut down unexpectedly", slog.Any("error", err))
	}
}

