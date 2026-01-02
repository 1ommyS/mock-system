package integration_test

import (
	"context"
	"testing"

	"user-svc/internal/infrastructure/postgres"
	"user-svc/tests/testutil"
)

func TestAuthzFlow(t *testing.T) {
	pg := testutil.StartPostgres(t)
	defer pg.Close(t)
	testutil.ApplyMigrations(t, pg.DB, "../../migrations")

	repos := postgres.NewStore(pg.DB)
	ctx := context.Background()

	admin, err := repos.Users.Create(ctx, "admin@example.com", "hash", "ACTIVE")
	if err != nil {
		t.Fatalf("create admin: %v", err)
	}
	user, err := repos.Users.Create(ctx, "user@example.com", "hash", "ACTIVE")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	if err := repos.Roles.ReplaceUserRoles(ctx, admin.ID, []string{"ADMIN", "USER"}); err != nil {
		t.Fatalf("assign admin roles: %v", err)
	}
	if err := repos.Roles.ReplaceUserRoles(ctx, user.ID, []string{"USER"}); err != nil {
		t.Fatalf("assign user roles: %v", err)
	}

	if _, err := repos.Sessions.Create(ctx, admin.ID, "refresh_admin", nil, nil); err != nil {
		t.Fatalf("create admin session: %v", err)
	}

	resID := "550e8400-e29b-41d4-a716-446655440000"
	res, created, err := repos.Resources.Register(ctx, "mock", resID, admin.ID)
	if err != nil {
		t.Fatalf("register resource: %v", err)
	}
	if !created {
		t.Fatalf("expected new resource, got existing")
	}
	if res.OwnerUserID != admin.ID {
		t.Fatalf("owner mismatch")
	}

	ac, err := repos.Resources.GetAccess(ctx, admin.ID, "mock", resID)
	if err != nil {
		t.Fatalf("access owner: %v", err)
	}
	if ac.EffectivePermission != "EDIT" || !ac.IsOwner {
		t.Fatalf("owner should have EDIT: %+v", ac)
	}

	grant, err := repos.Grants.Create(ctx, "mock", resID, user.ID, "READ")
	if err != nil {
		t.Fatalf("create grant: %v", err)
	}
	if grant.Permission != "READ" {
		t.Fatalf("grant permission mismatch")
	}

	acUser, err := repos.Resources.GetAccess(ctx, user.ID, "mock", resID)
	if err != nil {
		t.Fatalf("access user: %v", err)
	}
	if acUser.EffectivePermission != "READ" || acUser.IsOwner {
		t.Fatalf("user should have READ: %+v", acUser)
	}

	list, err := repos.Resources.ListByUser(ctx, user.ID, "mock", "read")
	if err != nil {
		t.Fatalf("list resources: %v", err)
	}
	if len(list) != 1 || list[0].ResourceID != resID {
		t.Fatalf("expected one resource, got %+v", list)
	}

	if _, err := repos.Grants.Create(ctx, "mock", resID, user.ID, "EDIT"); err != nil {
		if postgres.IsUniqueViolation(err) {
			if err := repos.Grants.Delete(ctx, "mock", resID, user.ID); err != nil {
				t.Fatalf("delete old grant: %v", err)
			}
			if _, err := repos.Grants.Create(ctx, "mock", resID, user.ID, "EDIT"); err != nil {
				t.Fatalf("recreate grant: %v", err)
			}
		} else {
			t.Fatalf("upgrade grant: %v", err)
		}
	}

	acUser, err = repos.Resources.GetAccess(ctx, user.ID, "mock", resID)
	if err != nil {
		t.Fatalf("access user after edit grant: %v", err)
	}
	if acUser.EffectivePermission != "EDIT" {
		t.Fatalf("expected EDIT after upgrade, got %+v", acUser)
	}

	if err := repos.Grants.Delete(ctx, "mock", resID, user.ID); err != nil {
		t.Fatalf("delete grant: %v", err)
	}
	acUser, err = repos.Resources.GetAccess(ctx, user.ID, "mock", resID)
	if err != nil {
		t.Fatalf("access user after revoke: %v", err)
	}
	if acUser.EffectivePermission != "NONE" || acUser.IsOwner {
		t.Fatalf("expected NONE after revoke, got %+v", acUser)
	}

	res, err = repos.Resources.UpdateOwner(ctx, "mock", resID, user.ID)
	if err != nil {
		t.Fatalf("update owner: %v", err)
	}
	if res.OwnerUserID != user.ID {
		t.Fatalf("owner not updated")
	}
	acUser, err = repos.Resources.GetAccess(ctx, user.ID, "mock", resID)
	if err != nil {
		t.Fatalf("access user after owner change: %v", err)
	}
	if acUser.EffectivePermission != "EDIT" || !acUser.IsOwner {
		t.Fatalf("new owner should have EDIT: %+v", acUser)
	}
}
