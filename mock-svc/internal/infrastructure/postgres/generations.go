package postgres

import (
	"context"
	"database/sql"
	"errors"

	"mock-svc/internal/domain"

	"github.com/jmoiron/sqlx"
)

type GenerationRepository struct {
	db *sqlx.DB
}

func NewGenerationRepository(db *sqlx.DB) *GenerationRepository {
	return &GenerationRepository{db: db}
}

func (r *GenerationRepository) Create(ctx context.Context, tx *sqlx.Tx, gen domain.Generation) (domain.Generation, error) {
	const q = `
		INSERT INTO generations (base_mock_id, dsl_script_id, created_by, status, params, result_mock_ids, error)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, base_mock_id, dsl_script_id, created_by, status, params, COALESCE(result_mock_ids, '{}') AS result_mock_ids, error, created_at, updated_at
	`
	var created domain.Generation
	if err := tx.GetContext(ctx, &created, q,
		gen.BaseMockID,
		gen.DSLScriptID,
		gen.CreatedBy,
		gen.Status,
		gen.Params,
		gen.ResultMockIDs,
		gen.Error,
	); err != nil {
		return domain.Generation{}, err
	}
	return created, nil
}

func (r *GenerationRepository) Get(ctx context.Context, id string) (domain.Generation, error) {
	const q = `
		SELECT id, base_mock_id, dsl_script_id, created_by, status, params, COALESCE(result_mock_ids, '{}') AS result_mock_ids, error, created_at, updated_at
		FROM generations
		WHERE id = $1
	`
	var gen domain.Generation
	if err := r.db.GetContext(ctx, &gen, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Generation{}, ErrNotFound
		}
		return domain.Generation{}, err
	}
	return gen, nil
}

func (r *GenerationRepository) UpdateStatus(ctx context.Context, tx *sqlx.Tx, id, status string, resultIDs []string, errMsg *string) error {
	const q = `
		UPDATE generations
		SET status = $2,
			result_mock_ids = $3,
			error = $4,
			updated_at = now()
		WHERE id = $1
	`
	res, err := tx.ExecContext(ctx, q, id, status, resultIDs, errMsg)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrNotFound
	}
	return nil
}
