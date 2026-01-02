package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type GrantRepository struct {
	db *sqlx.DB
}

func NewGrantRepository(db *sqlx.DB) *GrantRepository {
	return &GrantRepository{db: db}
}

func (r *GrantRepository) Create(ctx context.Context, resourceType, resourceID, userID, permission string) (Grant, error) {
	const q = `
		INSERT INTO resource_grants (resource_type, resource_id, user_id, permission)
		VALUES ($1, $2, $3, $4)
		RETURNING id, resource_type, resource_id, user_id, permission, created_at
	`
	var grant Grant
	if err := r.db.GetContext(ctx, &grant, q, resourceType, resourceID, userID, permission); err != nil {
		return Grant{}, err
	}
	return grant, nil
}

func (r *GrantRepository) Delete(ctx context.Context, resourceType, resourceID, userID string) error {
	const q = `DELETE FROM resource_grants WHERE resource_type = $1 AND resource_id = $2 AND user_id = $3`
	res, err := r.db.ExecContext(ctx, q, resourceType, resourceID, userID)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GrantRepository) ListByResource(ctx context.Context, resourceType, resourceID string) ([]Grant, error) {
	const q = `
		SELECT id, resource_type, resource_id, user_id, permission, created_at
		FROM resource_grants
		WHERE resource_type = $1 AND resource_id = $2
		ORDER BY created_at
	`
	var grants []Grant
	if err := r.db.SelectContext(ctx, &grants, q, resourceType, resourceID); err != nil {
		return nil, err
	}
	return grants, nil
}

func (r *GrantRepository) GetByUser(ctx context.Context, resourceType, resourceID, userID string) (Grant, error) {
	const q = `
		SELECT id, resource_type, resource_id, user_id, permission, created_at
		FROM resource_grants
		WHERE resource_type = $1 AND resource_id = $2 AND user_id = $3
	`
	var grant Grant
	if err := r.db.GetContext(ctx, &grant, q, resourceType, resourceID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Grant{}, ErrNotFound
		}
		return Grant{}, err
	}
	return grant, nil
}
