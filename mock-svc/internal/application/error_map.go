package application

import (
	"errors"

	"mock-svc/internal/domain"
	"mock-svc/internal/infrastructure/postgres"
)

func mapRepoError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, domain.ErrNotFound) {
		return ErrNotFound
	}
	if postgres.IsUniqueViolation(err) {
		return ErrConflict
	}
	if postgres.IsForeignKeyViolation(err) {
		return ErrNotFound
	}
	return err
}
