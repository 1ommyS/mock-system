package testutil

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PGContainer struct {
	Container testcontainers.Container
	DB        *sqlx.DB
}

func StartPostgres(t *testing.T) *PGContainer {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "dslrunner",
			"POSTGRES_USER":     "dsl",
			"POSTGRES_PASSWORD": "pass",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}

	pg, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}

	host, err := pg.Host(ctx)
	if err != nil {
		_ = pg.Terminate(ctx)
		t.Fatalf("host: %v", err)
	}

	port, err := pg.MappedPort(ctx, "5432/tcp")
	if err != nil {
		_ = pg.Terminate(ctx)
		t.Fatalf("mapped port: %v", err)
	}

	dsn := fmt.Sprintf("postgres://dsl:pass@%s:%s/dslrunner?sslmode=disable", host, port.Port())

	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		_ = pg.Terminate(ctx)
		t.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(15 * time.Minute)

	return &PGContainer{Container: pg, DB: db}
}

func (pg *PGContainer) Close(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if pg.DB != nil {
		_ = pg.DB.Close()
	}
	if pg.Container != nil {
		_ = pg.Container.Terminate(ctx)
	}
}

func ApplyMigrations(t *testing.T, db *sqlx.DB, migrationsDir string) {
	t.Helper()

	path := filepath.Join(migrationsDir, "0001_init.sql")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration %s: %v", path, err)
	}
	filtered := filterGooseDirectives(string(content))
	for _, stmt := range splitSQLStatements(filtered) {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("exec migration %s: %v", path, err)
		}
	}
}

func filterGooseDirectives(src string) string {
	var b strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(src))
	skip := false
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "-- +goose Down") {
			skip = true
			continue
		}
		if skip {
			continue
		}
		if strings.HasPrefix(trimmed, "-- +goose") {
			continue
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}

func splitSQLStatements(src string) []string {
	parts := strings.Split(src, ";")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		stmt := strings.TrimSpace(part)
		if stmt == "" {
			continue
		}
		out = append(out, stmt)
	}
	return out
}
