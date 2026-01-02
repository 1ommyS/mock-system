package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"
)

type ResourceRepository struct {
	db *sqlx.DB
}

func NewResourceRepository(db *sqlx.DB) *ResourceRepository {
	return &ResourceRepository{db: db}
}

func (r *ResourceRepository) Register(ctx context.Context, resourceType, resourceID, ownerUserID string) (Resource, bool, error) {
	const insertQ = `
		INSERT INTO resources (resource_type, resource_id, owner_user_id)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
		RETURNING resource_type, resource_id, owner_user_id, created_at
	`
	var res Resource
	if err := r.db.GetContext(ctx, &res, insertQ, resourceType, resourceID, ownerUserID); err == nil {
		return res, true, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Resource{}, false, err
	}

	const selectQ = `
		SELECT resource_type, resource_id, owner_user_id, created_at
		FROM resources
		WHERE resource_type = $1 AND resource_id = $2
	`
	if err := r.db.GetContext(ctx, &res, selectQ, resourceType, resourceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Resource{}, false, ErrNotFound
		}
		return Resource{}, false, err
	}

	return res, false, nil
}

func (r *ResourceRepository) Get(ctx context.Context, resourceType, resourceID string) (Resource, error) {
	const q = `
		SELECT resource_type, resource_id, owner_user_id, created_at
		FROM resources
		WHERE resource_type = $1 AND resource_id = $2
	`
	var res Resource
	if err := r.db.GetContext(ctx, &res, q, resourceType, resourceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Resource{}, ErrNotFound
		}
		return Resource{}, err
	}
	return res, nil
}

func (r *ResourceRepository) UpdateOwner(ctx context.Context, resourceType, resourceID, ownerUserID string) (Resource, error) {
	const q = `
		UPDATE resources
		SET owner_user_id = $3
		WHERE resource_type = $1 AND resource_id = $2
		RETURNING resource_type, resource_id, owner_user_id, created_at
	`
	var res Resource
	if err := r.db.GetContext(ctx, &res, q, resourceType, resourceID, ownerUserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Resource{}, ErrNotFound
		}
		return Resource{}, err
	}
	return res, nil
}

func (r *ResourceRepository) GetAccess(ctx context.Context, userID, resourceType, resourceID string) (AccessCheck, error) {
	const q = `
		SELECT
			CASE
				WHEN r.owner_user_id = $1 THEN 'EDIT'
				WHEN g.permission = 'EDIT' THEN 'EDIT'
				WHEN g.permission = 'READ' THEN 'READ'
				ELSE 'NONE'
			END AS effective_permission,
			(r.owner_user_id = $1) AS is_owner
		FROM resources r
		LEFT JOIN resource_grants g
			ON g.resource_type = r.resource_type
			AND g.resource_id = r.resource_id
			AND g.user_id = $1
		WHERE r.resource_type = $2 AND r.resource_id = $3
	`
	var res AccessCheck
	if err := r.db.GetContext(ctx, &res, q, userID, resourceType, resourceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AccessCheck{}, ErrNotFound
		}
		return AccessCheck{}, err
	}
	return res, nil
}

func (r *ResourceRepository) ListByUser(ctx context.Context, userID, resourceType, minPermission string) ([]ResourceAccess, error) {
	filter := ""
	switch strings.ToLower(minPermission) {
	case "read":
		filter = "AND (r.owner_user_id = $1 OR g.permission IN ('READ', 'EDIT'))"
	case "edit":
		filter = "AND (r.owner_user_id = $1 OR g.permission = 'EDIT')"
	case "":
		filter = ""
	default:
		filter = "AND false"
	}

	q := `
		SELECT
			r.resource_id,
			CASE
				WHEN r.owner_user_id = $1 THEN 'EDIT'
				WHEN g.permission = 'EDIT' THEN 'EDIT'
				WHEN g.permission = 'READ' THEN 'READ'
				ELSE 'NONE'
			END AS effective_permission,
			(r.owner_user_id = $1) AS is_owner
		FROM resources r
		LEFT JOIN resource_grants g
			ON g.resource_type = r.resource_type
			AND g.resource_id = r.resource_id
			AND g.user_id = $1
		WHERE r.resource_type = $2
		` + filter + `
		ORDER BY r.resource_id
	`

	var items []ResourceAccess
	if err := r.db.SelectContext(ctx, &items, q, userID, resourceType); err != nil {
		return nil, err
	}
	return items, nil
}
