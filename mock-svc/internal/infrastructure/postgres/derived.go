package postgres

import (
	"context"

	"mock-svc/internal/domain"

	"github.com/jmoiron/sqlx"
)

type DerivedRepository struct {
	db *sqlx.DB
}

func NewDerivedRepository(db *sqlx.DB) *DerivedRepository {
	return &DerivedRepository{db: db}
}

func (r *DerivedRepository) Create(ctx context.Context, tx *sqlx.Tx, rel domain.DerivedRelationship) error {
	const q = `
		INSERT INTO derived_relationships (base_mock_id, derived_mock_id, dsl_script_id, generation_id, params)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := tx.ExecContext(ctx, q, rel.BaseMockID, rel.DerivedMockID, rel.DSLScriptID, rel.GenerationID, rel.Params)
	return err
}

func (r *DerivedRepository) ListByDerivedIDs(ctx context.Context, derivedIDs []string) ([]domain.DerivedRelationship, error) {
	if len(derivedIDs) == 0 {
		return nil, nil
	}
	const q = `
		SELECT id, base_mock_id, derived_mock_id, dsl_script_id, generation_id, params, created_at
		FROM derived_relationships
		WHERE derived_mock_id = ANY($1)
	`
	var rels []domain.DerivedRelationship
	if err := r.db.SelectContext(ctx, &rels, q, derivedIDs); err != nil {
		return nil, err
	}
	return rels, nil
}
