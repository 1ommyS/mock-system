package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"mock-svc/internal/domain"

	"github.com/jmoiron/sqlx"
)

type MockRepository struct {
	db *sqlx.DB
}

func NewMockRepository(db *sqlx.DB) *MockRepository {
	return &MockRepository{db: db}
}

func (r *MockRepository) Create(ctx context.Context, tx *sqlx.Tx, mock domain.Mock) (domain.Mock, error) {
	const q = `
		INSERT INTO mocks (name, description, request_match, response_template, enabled, family_id, created_by, generated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, description, request_match, response_template, enabled, family_id, created_by, generated_by, created_at, updated_at, deleted_at
	`
	var created domain.Mock
	if err := tx.GetContext(ctx, &created, q,
		mock.Name,
		mock.Description,
		mock.RequestMatch,
		mock.ResponseTemplate,
		mock.Enabled,
		mock.FamilyID,
		mock.CreatedBy,
		mock.GeneratedBy,
	); err != nil {
		return domain.Mock{}, err
	}
	return created, nil
}

func (r *MockRepository) Get(ctx context.Context, id string) (domain.Mock, error) {
	const q = `
		SELECT id, name, description, request_match, response_template, enabled, family_id, created_by, generated_by, created_at, updated_at, deleted_at
		FROM mocks
		WHERE id = $1
	`
	var mock domain.Mock
	if err := r.db.GetContext(ctx, &mock, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Mock{}, ErrNotFound
		}
		return domain.Mock{}, err
	}
	return mock, nil
}

func (r *MockRepository) GetActive(ctx context.Context, id string) (domain.Mock, error) {
	const q = `
		SELECT id, name, description, request_match, response_template, enabled, family_id, created_by, generated_by, created_at, updated_at, deleted_at
		FROM mocks
		WHERE id = $1 AND deleted_at IS NULL
	`
	var mock domain.Mock
	if err := r.db.GetContext(ctx, &mock, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Mock{}, ErrNotFound
		}
		return domain.Mock{}, err
	}
	return mock, nil
}

func (r *MockRepository) Update(ctx context.Context, tx *sqlx.Tx, id string, patch domain.MockPatch) (domain.Mock, error) {
	set := make([]string, 0, 6)
	args := make([]any, 0, 7)
	idx := 1

	if patch.Name != nil {
		set = append(set, fmt.Sprintf("name = $%d", idx))
		args = append(args, *patch.Name)
		idx++
	}
	if patch.Description != nil {
		set = append(set, fmt.Sprintf("description = $%d", idx))
		args = append(args, *patch.Description)
		idx++
	}
	if patch.RequestMatch != nil {
		set = append(set, fmt.Sprintf("request_match = $%d", idx))
		args = append(args, *patch.RequestMatch)
		idx++
	}
	if patch.ResponseTemplate != nil {
		set = append(set, fmt.Sprintf("response_template = $%d", idx))
		args = append(args, *patch.ResponseTemplate)
		idx++
	}
	if patch.Enabled != nil {
		set = append(set, fmt.Sprintf("enabled = $%d", idx))
		args = append(args, *patch.Enabled)
		idx++
	}

	if len(set) == 0 {
		return domain.Mock{}, ErrNotFound
	}

	set = append(set, "updated_at = now()")
	args = append(args, id)

	q := fmt.Sprintf(`
		UPDATE mocks
		SET %s
		WHERE id = $%d AND deleted_at IS NULL
		RETURNING id, name, description, request_match, response_template, enabled, family_id, created_by, generated_by, created_at, updated_at, deleted_at
	`, strings.Join(set, ", "), idx)

	var updated domain.Mock
	if err := tx.GetContext(ctx, &updated, q, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Mock{}, ErrNotFound
		}
		return domain.Mock{}, err
	}
	return updated, nil
}

func (r *MockRepository) UpdateFamily(ctx context.Context, tx *sqlx.Tx, id string, familyID *string) (domain.Mock, error) {
	const q = `
		UPDATE mocks
		SET family_id = $2, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, name, description, request_match, response_template, enabled, family_id, created_by, generated_by, created_at, updated_at, deleted_at
	`
	var mock domain.Mock
	if err := tx.GetContext(ctx, &mock, q, id, familyID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Mock{}, ErrNotFound
		}
		return domain.Mock{}, err
	}
	return mock, nil
}

func (r *MockRepository) SoftDelete(ctx context.Context, tx *sqlx.Tx, id string) (domain.Mock, error) {
	const q = `
		UPDATE mocks
		SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, name, description, request_match, response_template, enabled, family_id, created_by, generated_by, created_at, updated_at, deleted_at
	`
	var mock domain.Mock
	if err := tx.GetContext(ctx, &mock, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Mock{}, ErrNotFound
		}
		return domain.Mock{}, err
	}
	return mock, nil
}

func (r *MockRepository) ListByIDs(ctx context.Context, ids []string) ([]domain.Mock, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	const q = `
		SELECT id, name, description, request_match, response_template, enabled, family_id, created_by, generated_by, created_at, updated_at, deleted_at
		FROM mocks
		WHERE id = ANY($1) AND deleted_at IS NULL
	`
	var items []domain.Mock
	if err := r.db.SelectContext(ctx, &items, q, ids); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MockRepository) ListByFamily(ctx context.Context, familyID string) ([]domain.Mock, error) {
	const q = `
		SELECT id, name, description, request_match, response_template, enabled, family_id, created_by, generated_by, created_at, updated_at, deleted_at
		FROM mocks
		WHERE family_id = $1 AND deleted_at IS NULL
	`
	var items []domain.Mock
	if err := r.db.SelectContext(ctx, &items, q, familyID); err != nil {
		return nil, err
	}
	return items, nil
}
