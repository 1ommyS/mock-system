package postgres

import (
	"context"
	"fmt"
	"strings"

	"mock-svc/internal/domain"

	"github.com/jmoiron/sqlx"
)

type MockSearchRepository struct {
	db *sqlx.DB
}

func NewMockSearchRepository(db *sqlx.DB) *MockSearchRepository {
	return &MockSearchRepository{db: db}
}

func (r *MockSearchRepository) Upsert(ctx context.Context, tx *sqlx.Tx, entry domain.MockSearch) error {
	const q = `
		INSERT INTO mock_search (mock_id, family_id, enabled, deleted_at, updated_at, method, path, name, description, search_tsv)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, to_tsvector('simple', coalesce($8, '') || ' ' || coalesce($9, '')))
		ON CONFLICT (mock_id) DO UPDATE
		SET family_id = EXCLUDED.family_id,
			enabled = EXCLUDED.enabled,
			deleted_at = EXCLUDED.deleted_at,
			updated_at = EXCLUDED.updated_at,
			method = EXCLUDED.method,
			path = EXCLUDED.path,
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			search_tsv = EXCLUDED.search_tsv
	`
	_, err := tx.ExecContext(ctx, q,
		entry.MockID,
		entry.FamilyID,
		entry.Enabled,
		entry.DeletedAt,
		entry.UpdatedAt,
		entry.Method,
		entry.Path,
		entry.Name,
		entry.Description,
	)
	return err
}

func (r *MockSearchRepository) Update(ctx context.Context, tx *sqlx.Tx, entry domain.MockSearch) error {
	const q = `
		UPDATE mock_search
		SET family_id = $2,
			enabled = $3,
			deleted_at = $4,
			updated_at = $5,
			method = $6,
			path = $7,
			name = $8,
			description = $9,
			search_tsv = to_tsvector('simple', coalesce($8, '') || ' ' || coalesce($9, ''))
		WHERE mock_id = $1
	`
	_, err := tx.ExecContext(ctx, q,
		entry.MockID,
		entry.FamilyID,
		entry.Enabled,
		entry.DeletedAt,
		entry.UpdatedAt,
		entry.Method,
		entry.Path,
		entry.Name,
		entry.Description,
	)
	return err
}

func (r *MockSearchRepository) Search(ctx context.Context, allowedIDs []string, filters domain.MockSearchFilters) ([]string, error) {
	if len(allowedIDs) == 0 {
		return nil, nil
	}
	clauses := []string{"deleted_at IS NULL", "mock_id = ANY($1)"}
	args := []any{allowedIDs}
	idx := 2

	if filters.FamilyID != nil {
		clauses = append(clauses, fmt.Sprintf("family_id = $%d", idx))
		args = append(args, *filters.FamilyID)
		idx++
	}
	if filters.Enabled != nil {
		clauses = append(clauses, fmt.Sprintf("enabled = $%d", idx))
		args = append(args, *filters.Enabled)
		idx++
	}
	if filters.Method != nil {
		clauses = append(clauses, fmt.Sprintf("method = $%d", idx))
		args = append(args, *filters.Method)
		idx++
	}
	if filters.Path != nil {
		clauses = append(clauses, fmt.Sprintf("path = $%d", idx))
		args = append(args, *filters.Path)
		idx++
	}
	if filters.Query != "" {
		clauses = append(clauses, fmt.Sprintf("search_tsv @@ plainto_tsquery('simple', $%d)", idx))
		args = append(args, filters.Query)
		idx++
	}

	limit := filters.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filters.Offset
	if offset < 0 {
		offset = 0
	}

	args = append(args, limit, offset)
	q := fmt.Sprintf(`
		SELECT mock_id
		FROM mock_search
		WHERE %s
		ORDER BY updated_at DESC
		LIMIT $%d OFFSET $%d
	`, strings.Join(clauses, " AND "), idx, idx+1)

	var ids []string
	if err := r.db.SelectContext(ctx, &ids, q, args...); err != nil {
		return nil, err
	}
	return ids, nil
}
