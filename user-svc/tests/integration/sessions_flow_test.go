package integration_test

import (
	"context"
	"testing"

	"user-svc/internal/infrastructure/postgres"
	"user-svc/tests/testutil"
)

func TestSessionsFlow(t *testing.T) {
	pg := testutil.StartPostgres(t)
	defer pg.Close(t)
	testutil.ApplyMigrations(t, pg.DB, "../../migrations")

	repos := postgres.NewStore(pg.DB)
	ctx := context.Background()

	user, err := repos.Users.Create(ctx, "login@example.com", "hash", "ACTIVE")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	sess, err := repos.Sessions.Create(ctx, user.ID, "refresh1", nil, nil)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if err := repos.Sessions.RotateRefresh(ctx, sess.ID, "refresh2"); err != nil {
		t.Fatalf("rotate: %v", err)
	}

	sess2, err := repos.Sessions.GetByRefreshHash(ctx, "refresh2")
	if err != nil {
		t.Fatalf("get by hash: %v", err)
	}
	if sess2.ID != sess.ID {
		t.Fatalf("session id mismatch")
	}

	if err := repos.Sessions.RevokeByID(ctx, sess.ID); err != nil {
		t.Fatalf("revoke: %v", err)
	}
}
