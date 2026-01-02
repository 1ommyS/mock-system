package integration_test

import (
	"context"
	"testing"
	"time"

	"user-svc/internal/infrastructure/postgres"
	"user-svc/tests/testutil"
)

func TestAuditLog(t *testing.T) {
	pg := testutil.StartPostgres(t)
	defer pg.Close(t)
	testutil.ApplyMigrations(t, pg.DB, "../../migrations")

	repos := postgres.NewStore(pg.DB)
	ctx := context.Background()

	entry := postgres.AuditLog{
		ActorUserID: nil,
		EventType:   "LOGIN",
		Metadata:    []byte(`{"ok":true}`),
	}
	out, err := repos.Audit.Create(ctx, entry)
	if err != nil {
		t.Fatalf("audit create: %v", err)
	}
	if out.EventType != "LOGIN" {
		t.Fatalf("event mismatch")
	}
	if time.Since(out.CreatedAt) > time.Minute {
		t.Fatalf("created_at not set")
	}
}
