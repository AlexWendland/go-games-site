package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/rs/cors"

	"github.com/AlexWendland/go-games-site/internal/api"
	"github.com/AlexWendland/go-games-site/internal/config"
	"github.com/AlexWendland/go-games-site/internal/db"
	"github.com/AlexWendland/go-games-site/internal/db/migrations"
	"github.com/AlexWendland/go-games-site/internal/middleware"
	"github.com/AlexWendland/go-games-site/internal/service"
	"github.com/AlexWendland/go-games-site/internal/web"
)

func main() {
	// Set up logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Read environment variables
	cfg := config.Load()

	// Set up database
	dbConnection, err := db.Open(cfg.DatabaseFile, migrations.FS)
	if err != nil {
		logger.Error("can not open database connection", slog.Any("error", err))
		return
	}
	defer dbConnection.Close()

	// Set up services
	authService := service.MakeAuthService(dbConnection)
	userService := service.MakeUserService(dbConnection)

	// Set up CORS
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})

	// Set up server
	server := http.NewServeMux()

	// Add handlers here
	server.Handle("/", web.WebServerHandler())
	server.Handle("/api/", corsMiddleware.Handler(
		http.StripPrefix("/api", api.ApiHandler(&cfg, authService, userService, middleware.Auth(authService))),
	))

	addr := fmt.Sprintf(":%d", cfg.Port)
	logger.Info("hosting server", "addr", addr)
	if err := http.ListenAndServe(addr, server); err != nil {
		logger.Error("server shut down unexpectedly", slog.Any("error", err))
	}
}
