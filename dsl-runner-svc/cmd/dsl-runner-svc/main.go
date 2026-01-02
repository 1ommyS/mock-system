package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dsl-runner-svc/internal/application"
	"dsl-runner-svc/internal/config"
	"dsl-runner-svc/internal/db"
	"dsl-runner-svc/internal/dsl"
	httpserver "dsl-runner-svc/internal/http"
	"dsl-runner-svc/internal/http/handlers"
	"dsl-runner-svc/internal/infrastructure/postgres"
	"dsl-runner-svc/internal/kafka"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	level, err := config.ParseLogLevel(cfg.LogLevel)
	if err != nil {
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
	executor := dsl.NewExecutor()
	publisher := kafka.NewResultPublisher(cfg.KafkaBrokers, cfg.KafkaResultsTopic)
	defer publisher.Writer.Close()

	repos := application.Repositories{Jobs: store.Jobs, Results: store.Results}
	service := application.New(repos, store, publisher, executor, cfg.WorkerMaxAttempts)
	handler := handlers.New(service)
	router := httpserver.NewRouter(handler, cfg.InternalSecret, cfg.InternalAuthHeaderName)

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: router, ReadHeaderTimeout: 5 * time.Second}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// start HTTP server
	go func() {
		logger.Info("dsl-runner-svc listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", "error", err)
			stop()
		}
	}()

	// start kafka consumers
	consumerCtx, cancelConsumers := context.WithCancel(ctx)
	for i := 0; i < cfg.WorkerConcurrency; i++ {
		reader := kafka.NewReader(cfg.KafkaBrokers, cfg.KafkaJobsTopic, cfg.KafkaGroupID)
		c := &kafka.Consumer{Reader: reader, Handler: &application.JobHandler{Service: service}, Logger: logger}
		go func(idx int) {
			logger.Info("worker started", "worker", idx)
			if err := c.Run(consumerCtx); err != nil {
				logger.Error("worker stopped", "worker", idx, "error", err)
				cancelConsumers()
			}
			_ = reader.Close()
		}(i)
	}

	<-ctx.Done()
	logger.Info("shutdown initiated")
	_ = srv.Shutdown(context.Background())
	cancelConsumers()
	_ = publisher.Writer.Close()
}
