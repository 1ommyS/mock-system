package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mock-svc/internal/application"
	"mock-svc/internal/config"
	"mock-svc/internal/db"
	httpserver "mock-svc/internal/http"
	"mock-svc/internal/http/handlers"
	"mock-svc/internal/http/middleware"
	"mock-svc/internal/infrastructure/postgres"
	"mock-svc/internal/service"
)

// @title Mock Storage Service API
// @version 1.0
// @description Mock storage service
// @BasePath /mocks/v1
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	database, err := db.Open(cfg.DBDSN)
	if err != nil {
		logger.Error("db connection failed", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	store := postgres.NewStore(database)
	repos := application.Repositories{
		Mocks:       store.Mocks,
		Families:    store.Families,
		Derived:     store.Derived,
		Generations: store.Generations,
		MockSearch:  store.MockSearch,
		Outbox:      store.Outbox,
	}
	authClient := service.NewAuthClient(cfg.AuthBaseURL, cfg.AuthInternalSecret)
	kafkaPublisher := service.NewKafkaPublisher(cfg.KafkaBrokers, cfg.KafkaJobsTopic)
	defer kafkaPublisher.Writer.Close()
	dslClient := service.NewDSLRunnerClient(service.DSLRunnerConfig{
		BaseURL:        cfg.DSLRunnerBaseURL,
		InternalSecret: cfg.DSLRunnerInternalSecret,
		PollInterval:   cfg.DSLRunnerPollInterval,
		PollTimeout:    cfg.DSLRunnerPollTimeout,
		JobsTopic:      cfg.KafkaJobsTopic,
		Limits: service.JobLimits{
			TimeoutMs:         cfg.DSLRunnerTimeoutMs,
			MaxGeneratedMocks: cfg.DSLRunnerMaxGenerated,
			MaxResultBytes:    cfg.DSLRunnerMaxResultBytes,
		},
		Publisher: kafkaPublisher,
	})
	services := application.New(repos, store, authClient, dslClient, cfg.OutboxMaxAttempt)
	handler := handlers.New(services)
	router := httpserver.NewRouter(handler, middleware.CORSConfig{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-User-Id", "X-Roles"},
		AllowCredentials: cfg.CORSAllowCredentials,
	})

	outboxWorker := &application.OutboxWorker{
		Repo:      store.Outbox,
		Auth:      authClient,
		Interval:  cfg.OutboxInterval,
		BatchSize: cfg.OutboxBatchSize,
	}
	workerCtx, cancelWorker := context.WithCancel(context.Background())
	go outboxWorker.Run(workerCtx)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("mock-svc listening", "addr", cfg.HTTPAddr, "log_level", cfg.LogLevel)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	cancelWorker()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", "error", err)
	}
}
