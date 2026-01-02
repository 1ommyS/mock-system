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

func TestAuditRepository_Create(t *testing.T) {
	db, mock, cleanup := testutil.NewMockDB(t)
	defer cleanup()
	repo := postgres.NewAuditRepository(db)
	ctx := context.Background()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "actor_user_id", "event_type", "resource_type", "resource_id", "ip", "user_agent", "metadata", "created_at",
	}).AddRow("audit-1", "user-1", "LOGIN", nil, nil, "127.0.0.1", "ua", []byte(`{}`), now)

	mock.ExpectQuery(`INSERT INTO audit_log .* RETURNING`).
		WithArgs(sqlmock.AnyArg(), "LOGIN", nil, nil, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	entry := postgres.AuditLog{
		ActorUserID: testutil.Ptr("user-1"),
		EventType:   "LOGIN",
		IP:          testutil.Ptr("127.0.0.1"),
		UserAgent:   testutil.Ptr("ua"),
		Metadata:    []byte(`{}`),
	}

	out, err := repo.Create(ctx, entry)
	require.NoError(t, err)
	require.Equal(t, "audit-1", out.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}
