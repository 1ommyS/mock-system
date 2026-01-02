package integration_test

import (
	"context"
	"testing"

	"user-svc/internal/infrastructure/postgres"
	"user-svc/tests/testutil"
)

func TestResourceRegisterIdempotentOwnerImmutable(t *testing.T) {
	pg := testutil.StartPostgres(t)
	defer pg.Close(t)
	testutil.ApplyMigrations(t, pg.DB, "../../migrations")

	repos := postgres.NewStore(pg.DB)
	ctx := context.Background()

	owner1, err := repos.Users.Create(ctx, "owner1@example.com", "hash", "ACTIVE")
	if err != nil {
		t.Fatalf("create owner1: %v", err)
	}
	owner2, err := repos.Users.Create(ctx, "owner2@example.com", "hash", "ACTIVE")
	if err != nil {
		t.Fatalf("create owner2: %v", err)
	}

	resID := "550e8400-e29b-41d4-a716-446655440001"
	_, created, err := repos.Resources.Register(ctx, "mock", resID, owner1.ID)
	if err != nil {
		t.Fatalf("register first: %v", err)
	}
	if !created {
		t.Fatalf("expected created")
	}

	res2, created2, err := repos.Resources.Register(ctx, "mock", resID, owner2.ID)
	if err != nil {
		t.Fatalf("register second: %v", err)
	}
	if created2 {
		t.Fatalf("expected existing resource on second register")
	}
	if res2.OwnerUserID != owner1.ID {
		t.Fatalf("owner should remain owner1, got %s", res2.OwnerUserID)
	}
}

func TestListResourcesByPermission(t *testing.T) {
	pg := testutil.StartPostgres(t)
	defer pg.Close(t)
	testutil.ApplyMigrations(t, pg.DB, "../../migrations")

	repos := postgres.NewStore(pg.DB)
	ctx := context.Background()

	ownerA, _ := repos.Users.Create(ctx, "ownera@example.com", "hash", "ACTIVE")
	ownerB, _ := repos.Users.Create(ctx, "ownerb@example.com", "hash", "ACTIVE")
	userX, _ := repos.Users.Create(ctx, "userx@example.com", "hash", "ACTIVE")

	resA := "550e8400-e29b-41d4-a716-446655440010"
	resB := "550e8400-e29b-41d4-a716-446655440011"

	if _, _, err := repos.Resources.Register(ctx, "mock", resA, ownerA.ID); err != nil {
		t.Fatalf("register resA: %v", err)
	}
	if _, _, err := repos.Resources.Register(ctx, "mock", resB, ownerB.ID); err != nil {
		t.Fatalf("register resB: %v", err)
	}

	if _, err := repos.Grants.Create(ctx, "mock", resA, userX.ID, "READ"); err != nil {
		t.Fatalf("grant read: %v", err)
	}
	if _, err := repos.Grants.Create(ctx, "mock", resB, userX.ID, "EDIT"); err != nil {
		t.Fatalf("grant edit: %v", err)
	}

	readList, err := repos.Resources.ListByUser(ctx, userX.ID, "mock", "read")
	if err != nil {
		t.Fatalf("list read: %v", err)
	}
	if len(readList) != 2 {
		t.Fatalf("expected 2 resources for read, got %d", len(readList))
	}

	editList, err := repos.Resources.ListByUser(ctx, userX.ID, "mock", "edit")
	if err != nil {
		t.Fatalf("list edit: %v", err)
	}
	if len(editList) != 1 || editList[0].ResourceID != resB {
		t.Fatalf("expected only resB for edit, got %+v", editList)
	}
}
