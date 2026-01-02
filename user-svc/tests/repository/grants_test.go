package repository_test

import (
	"context"
	"testing"
	"time"

	"user-svc/internal/infrastructure/postgres"
	"user-svc/tests/testutil"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestGrantRepository_Create(t *testing.T) {
	db, mock, cleanup := testutil.NewMockDB(t)
	defer cleanup()
	repo := postgres.NewGrantRepository(db)
	ctx := context.Background()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "resource_type", "resource_id", "user_id", "permission", "created_at",
	}).AddRow("grant-1", "mock", "res-1", "user-1", "READ", now)

	mock.ExpectQuery(`INSERT INTO resource_grants .* RETURNING`).
		WithArgs("mock", "res-1", "user-1", "READ").
		WillReturnRows(rows)

	grant, err := repo.Create(ctx, "mock", "res-1", "user-1", "READ")
	require.NoError(t, err)
	require.Equal(t, "grant-1", grant.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGrantRepository_Delete_NotFound(t *testing.T) {
	db, mock, cleanup := testutil.NewMockDB(t)
	defer cleanup()
	repo := postgres.NewGrantRepository(db)

	mock.ExpectExec(`DELETE FROM resource_grants WHERE resource_type = \$1 AND resource_id = \$2 AND user_id = \$3`).
		WithArgs("mock", "res-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), "mock", "res-1", "user-1")
	require.ErrorIs(t, err, postgres.ErrNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGrantRepository_ListByResource(t *testing.T) {
	db, mock, cleanup := testutil.NewMockDB(t)
	defer cleanup()
	repo := postgres.NewGrantRepository(db)
	ctx := context.Background()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "resource_type", "resource_id", "user_id", "permission", "created_at",
	}).
		AddRow("grant-1", "mock", "res-1", "user-1", "READ", now).
		AddRow("grant-2", "mock", "res-1", "user-2", "EDIT", now)

	mock.ExpectQuery(`SELECT id, resource_type, resource_id, user_id, permission, created_at FROM resource_grants WHERE resource_type = \$1 AND resource_id = \$2 ORDER BY created_at`).
		WithArgs("mock", "res-1").
		WillReturnRows(rows)

	grants, err := repo.ListByResource(ctx, "mock", "res-1")
	require.NoError(t, err)
	require.Len(t, grants, 2)
	require.Equal(t, "READ", grants[0].Permission)
	require.NoError(t, mock.ExpectationsWereMet())
}
