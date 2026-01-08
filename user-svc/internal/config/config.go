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
	HTTPAddr             string        `yaml:"http_addr"`
	DBDSN                string        `yaml:"db_dsn"`
	JWTSecret            string        `yaml:"jwt_secret"`
	AccessTTL            time.Duration `yaml:"access_ttl"`
	RefreshTTL           time.Duration `yaml:"refresh_ttl"`
	Issuer               string        `yaml:"issuer"`
	Audience             string        `yaml:"audience"`
	LogLevel             string        `yaml:"log_level"`
	CORSAllowedOrigins   []string      `yaml:"cors_allowed_origins"`
	CORSAllowCredentials bool          `yaml:"cors_allow_credentials"`
	AuthInternalSecret   string        `yaml:"auth_internal_secret"`
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
		cfg.HTTPAddr = ":8080"
	}
	if cfg.Issuer == "" {
		cfg.Issuer = "user-svc"
	}
	if cfg.Audience == "" {
		cfg.Audience = "user-svc"
	}
	if cfg.AccessTTL == 0 {
		cfg.AccessTTL = 15 * time.Minute
	}
	if cfg.RefreshTTL == 0 {
		cfg.RefreshTTL = 30 * 24 * time.Hour
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if len(cfg.CORSAllowedOrigins) == 0 {
		cfg.CORSAllowedOrigins = []string{"http://localhost:3000"}
	}
	if cfg.AuthInternalSecret == "" {
		cfg.AuthInternalSecret = "dev-secret"
	}

	if cfg.DBDSN == "" {
		return Config{}, fmt.Errorf("db_dsn is required")
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("jwt_secret is required")
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
