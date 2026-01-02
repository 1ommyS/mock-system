package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"user-svc/internal/application"
	"user-svc/internal/auth"
	"user-svc/internal/config"
	"user-svc/internal/db"
	httpserver "user-svc/internal/http"
	"user-svc/internal/http/handlers"
	"user-svc/internal/infrastructure/postgres"
)

// @title User Service Auth & ACL API
// @version 1.0
// @description Authentication and ACL service
// @BasePath /
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
//
//go:generate swag init -g cmd/user-svc/main.go -o internal/docs --outputTypes yaml,json
func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	level, err := config.ParseLogLevel(cfg.LogLevel)
	if err != nil {
		slog.Warn("invalid log level, falling back to info", "log_level", cfg.LogLevel)
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	database, err := db.Open(cfg.DBDSN)
	if err != nil {
		logger.Error("db connection failed", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	jwtSvc := auth.NewJWTService(cfg.JWTSecret, cfg.Issuer, cfg.Audience)
	store := postgres.NewStore(database)
	repos := application.Repositories{
		Users:     store.Users,
		Roles:     store.Roles,
		Sessions:  store.Sessions,
		Resources: store.Resources,
		Grants:    store.Grants,
		Audit:     store.Audit,
	}
	services := application.New(repos, jwtSvc, cfg.AccessTTL, cfg.RefreshTTL)
	handler := handlers.New(services)
	router := httpserver.NewRouter(handler, jwtSvc)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("user-svc listening", "addr", cfg.HTTPAddr, "log_level", cfg.LogLevel)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", "error", err)
	}
}
