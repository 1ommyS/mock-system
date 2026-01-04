package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTPAddr                string        `yaml:"http_addr"`
	DBDSN                   string        `yaml:"db_dsn"`
	LogLevel                string        `yaml:"log_level"`
	AuthBaseURL             string        `yaml:"auth_base_url"`
	DSLRunnerBaseURL        string        `yaml:"dsl_runner_base_url"`
	DSLRunnerInternalSecret string        `yaml:"dsl_runner_internal_secret"`
	DSLRunnerPollInterval   time.Duration `yaml:"dsl_runner_poll_interval"`
	DSLRunnerPollTimeout    time.Duration `yaml:"dsl_runner_poll_timeout"`
	DSLRunnerTimeoutMs      int           `yaml:"dsl_runner_timeout_ms"`
	DSLRunnerMaxGenerated   int           `yaml:"dsl_runner_max_generated_mocks"`
	DSLRunnerMaxResultBytes int           `yaml:"dsl_runner_max_result_bytes"`
	KafkaBrokers            []string      `yaml:"kafka_brokers"`
	KafkaJobsTopic          string        `yaml:"kafka_jobs_topic"`
	OutboxInterval          time.Duration `yaml:"outbox_interval"`
	OutboxBatchSize         int           `yaml:"outbox_batch_size"`
	OutboxMaxAttempt        int           `yaml:"outbox_max_attempts"`
}

func Load() (Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "config.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}

	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8081"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.OutboxInterval == 0 {
		cfg.OutboxInterval = 2 * time.Second
	}
	if cfg.OutboxBatchSize == 0 {
		cfg.OutboxBatchSize = 50
	}
	if cfg.OutboxMaxAttempt == 0 {
		cfg.OutboxMaxAttempt = 10
	}
	if cfg.DSLRunnerPollInterval == 0 {
		cfg.DSLRunnerPollInterval = 200 * time.Millisecond
	}
	if cfg.DSLRunnerPollTimeout == 0 {
		cfg.DSLRunnerPollTimeout = 5 * time.Second
	}
	if cfg.DSLRunnerTimeoutMs == 0 {
		cfg.DSLRunnerTimeoutMs = 5000
	}
	if cfg.DSLRunnerMaxGenerated == 0 {
		cfg.DSLRunnerMaxGenerated = 200
	}
	if cfg.DSLRunnerMaxResultBytes == 0 {
		cfg.DSLRunnerMaxResultBytes = 2_000_000
	}
	if cfg.KafkaJobsTopic == "" {
		cfg.KafkaJobsTopic = "dsl.jobs.v1"
	}

	if cfg.DBDSN == "" {
		return Config{}, fmt.Errorf("db_dsn is required")
	}
	if cfg.AuthBaseURL == "" {
		return Config{}, fmt.Errorf("auth_base_url is required")
	}
	if cfg.DSLRunnerBaseURL == "" {
		return Config{}, fmt.Errorf("dsl_runner_base_url is required")
	}
	if cfg.DSLRunnerInternalSecret == "" {
		return Config{}, fmt.Errorf("dsl_runner_internal_secret is required")
	}
	if len(cfg.KafkaBrokers) == 0 {
		return Config{}, fmt.Errorf("kafka_brokers is required")
	}

	return cfg, nil
}

func ParseLogLevel(lvl string) (slog.Level, error) {
	switch strings.ToLower(lvl) {
	case "debug":
		return slog.LevelDebug, nil
	case "info", "":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown log_level: %s", lvl)
	}
}
