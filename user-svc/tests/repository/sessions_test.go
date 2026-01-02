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

func TestSessionRepository_Create(t *testing.T) {
	db, mock, cleanup := testutil.NewMockDB(t)
	defer cleanup()
	repo := postgres.NewSessionRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "refresh_hash", "created_at", "last_used_at", "revoked_at", "ip", "user_agent",
	}).AddRow("sess-1", "user-1", "hash", now, now, nil, "127.0.0.1", "ua")

	mock.ExpectQuery(`INSERT INTO sessions .* RETURNING`).
		WithArgs("user-1", "hash", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	ip := "127.0.0.1"
	ua := "ua"
	session, err := repo.Create(context.Background(), "user-1", "hash", &ip, &ua)
	require.NoError(t, err)
	require.Equal(t, "sess-1", session.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_RevokeByRefreshHash_NotFound(t *testing.T) {
	db, mock, cleanup := testutil.NewMockDB(t)
	defer cleanup()
	repo := postgres.NewSessionRepository(db)

	mock.ExpectExec(`UPDATE sessions SET revoked_at = now\(\) WHERE refresh_hash = \$1 AND revoked_at IS NULL`).
		WithArgs("hash").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.RevokeByRefreshHash(context.Background(), "hash")
	require.ErrorIs(t, err, postgres.ErrNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}
