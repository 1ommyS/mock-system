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
	HTTPAddr               string        `yaml:"http_addr"`
	DBDSN                  string        `yaml:"db_dsn"`
	LogLevel               string        `yaml:"log_level"`
	InternalSecret         string        `yaml:"internal_secret"`
	KafkaBrokers           []string      `yaml:"kafka_brokers"`
	KafkaJobsTopic         string        `yaml:"kafka_jobs_topic"`
	KafkaResultsTopic      string        `yaml:"kafka_results_topic"`
	KafkaGroupID           string        `yaml:"kafka_group_id"`
	WorkerConcurrency      int           `yaml:"worker_concurrency"`
	WorkerMaxAttempts      int           `yaml:"worker_max_attempts"`
	WorkerTimeout          time.Duration `yaml:"worker_timeout"`
	WorkerMaxGenerated     int           `yaml:"worker_max_generated_mocks"`
	WorkerMaxResultBytes   int           `yaml:"worker_max_result_bytes"`
	InternalAuthHeaderName string        `yaml:"internal_auth_header_name"`
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
		cfg.HTTPAddr = ":8090"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.KafkaJobsTopic == "" {
		cfg.KafkaJobsTopic = "dsl.jobs.v1"
	}
	if cfg.KafkaResultsTopic == "" {
		cfg.KafkaResultsTopic = "dsl.results.v1"
	}
	if cfg.KafkaGroupID == "" {
		cfg.KafkaGroupID = "dsl-runner-workers"
	}
	if cfg.WorkerConcurrency == 0 {
		cfg.WorkerConcurrency = 8
	}
	if cfg.WorkerMaxAttempts == 0 {
		cfg.WorkerMaxAttempts = 5
	}
	if cfg.WorkerTimeout == 0 {
		cfg.WorkerTimeout = 5 * time.Second
	}
	if cfg.WorkerMaxGenerated == 0 {
		cfg.WorkerMaxGenerated = 200
	}
	if cfg.WorkerMaxResultBytes == 0 {
		cfg.WorkerMaxResultBytes = 2_000_000
	}
	if cfg.InternalAuthHeaderName == "" {
		cfg.InternalAuthHeaderName = "X-Internal-Secret"
	}

	if cfg.DBDSN == "" {
		return Config{}, fmt.Errorf("db_dsn is required")
	}
	if cfg.InternalSecret == "" {
		return Config{}, fmt.Errorf("internal_secret is required")
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
