package postgres

import (
	"errors"

	"mock-svc/internal/domain"

	"github.com/jackc/pgx/v5/pgconn"
)

var ErrNotFound = domain.ErrNotFound

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func IsForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}
	return false
}
