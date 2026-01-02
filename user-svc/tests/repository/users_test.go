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

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	db, mock, cleanup := testutil.NewMockDB(t)
	defer cleanup()

	repo := postgres.NewUserRepository(db)

	mock.ExpectQuery(`SELECT .* FROM users WHERE lower\(email\) = lower\(\$1\)`).
		WithArgs("missing@example.com").
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByEmail(context.Background(), "missing@example.com")
	require.ErrorIs(t, err, postgres.ErrNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create(t *testing.T) {
	db, mock, cleanup := testutil.NewMockDB(t)
	defer cleanup()

	repo := postgres.NewUserRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "email", "password_hash", "status", "created_at", "updated_at", "last_login_at",
	}).AddRow("user-1", "user@example.com", "hash", "ACTIVE", now, now, nil)

	mock.ExpectQuery(`INSERT INTO users .* RETURNING`).
		WithArgs("user@example.com", "hash", "ACTIVE").
		WillReturnRows(rows)

	user, err := repo.Create(context.Background(), "USER@EXAMPLE.COM", "hash", "ACTIVE")
	require.NoError(t, err)
	require.Equal(t, "user-1", user.ID)
	require.Equal(t, "user@example.com", user.Email)
	require.NoError(t, mock.ExpectationsWereMet())
}
