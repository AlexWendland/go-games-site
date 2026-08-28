package main

import (
	"net/http"
	"log/slog"
	"os"

	"github.com/AlexWendland/go-games-site/internal/web"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	server := http.NewServeMux()

	// Add handlers here
	server.Handle("/", web.NewHandler())

	logger.Info("hosting server on http://localhost:8080/")
	if err := http.ListenAndServe(":8080", server); err != nil {
		logger.Error("server shut down unexpectedly", slog.Any("error", err))
	}
}

