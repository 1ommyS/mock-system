package repository_test

import (
	"context"
	"testing"

	"user-svc/internal/infrastructure/postgres"
	"user-svc/tests/testutil"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestRoleRepository_ReplaceUserRoles(t *testing.T) {
	db, mock, cleanup := testutil.NewMockDB(t)
	defer cleanup()

	repo := postgres.NewRoleRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO roles \(name\) VALUES \(\$1\) ON CONFLICT DO NOTHING`).
		WithArgs("ADMIN").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO roles \(name\) VALUES \(\$1\) ON CONFLICT DO NOTHING`).
		WithArgs("USER").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM user_roles WHERE user_id = \$1`).
		WithArgs("user-1").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO user_roles \(user_id, role_name\) VALUES \(\$1, \$2\), \(\$3, \$4\)`).
		WithArgs("user-1", "ADMIN", "user-1", "USER").
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	err := repo.ReplaceUserRoles(ctx, "user-1", []string{"ADMIN", "USER"})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
