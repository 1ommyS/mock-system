package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"user-svc/internal/infrastructure/postgres"
	"user-svc/tests/testutil"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestResourceRepository_Register_New(t *testing.T) {
	db, mock, cleanup := testutil.NewMockDB(t)
	defer cleanup()
	repo := postgres.NewResourceRepository(db)
	ctx := context.Background()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"resource_type", "resource_id", "owner_user_id", "created_at",
	}).AddRow("mock", "res-1", "owner-1", now)

	mock.ExpectQuery(`INSERT INTO resources .* RETURNING`).
		WithArgs("mock", "res-1", "owner-1").
		WillReturnRows(rows)

	res, created, err := repo.Register(ctx, "mock", "res-1", "owner-1")
	require.NoError(t, err)
	require.True(t, created)
	require.Equal(t, "res-1", res.ResourceID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestResourceRepository_Register_Existing(t *testing.T) {
	db, mock, cleanup := testutil.NewMockDB(t)
	defer cleanup()
	repo := postgres.NewResourceRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`INSERT INTO resources .* RETURNING`).
		WithArgs("mock", "res-1", "owner-1").
		WillReturnError(sql.ErrNoRows)

	now := time.Now()
	existing := sqlmock.NewRows([]string{
		"resource_type", "resource_id", "owner_user_id", "created_at",
	}).AddRow("mock", "res-1", "owner-1", now)

	mock.ExpectQuery(`SELECT resource_type, resource_id, owner_user_id, created_at FROM resources WHERE resource_type = \$1 AND resource_id = \$2`).
		WithArgs("mock", "res-1").
		WillReturnRows(existing)

	res, created, err := repo.Register(ctx, "mock", "res-1", "owner-1")
	require.NoError(t, err)
	require.False(t, created)
	require.Equal(t, "res-1", res.ResourceID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestResourceRepository_GetAccess(t *testing.T) {
	db, mock, cleanup := testutil.NewMockDB(t)
	defer cleanup()
	repo := postgres.NewResourceRepository(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"effective_permission", "is_owner",
	}).AddRow("EDIT", true)

	mock.ExpectQuery(`SELECT\s+CASE\s+.*FROM resources r\s+LEFT JOIN resource_grants g`).
		WithArgs("user-1", "mock", "res-1").
		WillReturnRows(rows)

	ac, err := repo.GetAccess(ctx, "user-1", "mock", "res-1")
	require.NoError(t, err)
	require.Equal(t, "EDIT", ac.EffectivePermission)
	require.True(t, ac.IsOwner)
	require.NoError(t, mock.ExpectationsWereMet())
}
