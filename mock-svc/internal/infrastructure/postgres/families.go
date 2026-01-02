package postgres

import (
	"context"
	"database/sql"
	"errors"

	"mock-svc/internal/domain"

	"github.com/jmoiron/sqlx"
)

type FamilyRepository struct {
	db *sqlx.DB
}

func NewFamilyRepository(db *sqlx.DB) *FamilyRepository {
	return &FamilyRepository{db: db}
}

func (r *FamilyRepository) Create(ctx context.Context, tx *sqlx.Tx, primaryMockID string) (domain.Family, error) {
	const q = `
		INSERT INTO families (primary_mock_id)
		VALUES ($1)
		RETURNING id, primary_mock_id
	`
	var family domain.Family
	if err := tx.GetContext(ctx, &family, q, primaryMockID); err != nil {
		return domain.Family{}, err
	}
	return family, nil
}

func (r *FamilyRepository) Get(ctx context.Context, id string) (domain.Family, error) {
	const q = `
		SELECT id, primary_mock_id
		FROM families
		WHERE id = $1
	`
	var family domain.Family
	if err := r.db.GetContext(ctx, &family, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Family{}, ErrNotFound
		}
		return domain.Family{}, err
	}
	return family, nil
}

func (r *FamilyRepository) UpdatePrimary(ctx context.Context, tx *sqlx.Tx, id, newPrimaryID string) (domain.Family, error) {
	const q = `
		UPDATE families
		SET primary_mock_id = $2
		WHERE id = $1
		RETURNING id, primary_mock_id
	`
	var family domain.Family
	if err := tx.GetContext(ctx, &family, q, id, newPrimaryID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Family{}, ErrNotFound
		}
		return domain.Family{}, err
	}
	return family, nil
}
