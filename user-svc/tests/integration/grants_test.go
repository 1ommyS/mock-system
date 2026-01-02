package integration_test

import (
	"context"
	"testing"

	"user-svc/internal/infrastructure/postgres"
	"user-svc/tests/testutil"
)

func TestGrantUniqueViolation(t *testing.T) {
	pg := testutil.StartPostgres(t)
	defer pg.Close(t)
	testutil.ApplyMigrations(t, pg.DB, "../../migrations")

	repos := postgres.NewStore(pg.DB)
	ctx := context.Background()

	owner, _ := repos.Users.Create(ctx, "owner@example.com", "hash", "ACTIVE")
	user, _ := repos.Users.Create(ctx, "user@example.com", "hash", "ACTIVE")

	resID := "550e8400-e29b-41d4-a716-446655440020"
	if _, _, err := repos.Resources.Register(ctx, "mock", resID, owner.ID); err != nil {
		t.Fatalf("register: %v", err)
	}

	if _, err := repos.Grants.Create(ctx, "mock", resID, user.ID, "READ"); err != nil {
		t.Fatalf("grant: %v", err)
	}

	if _, err := repos.Grants.Create(ctx, "mock", resID, user.ID, "READ"); err == nil {
		t.Fatalf("expected unique violation on duplicate grant")
	} else if !postgres.IsUniqueViolation(err) {
		t.Fatalf("expected unique violation, got %v", err)
	}
}
