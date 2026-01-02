package postgres

import (
	"context"
	"database/sql"

	"dsl-runner-svc/internal/domain"

	"github.com/jmoiron/sqlx"
)

type JobResultRepository struct {
	db *sqlx.DB
}

func NewJobResultRepository(db *sqlx.DB) *JobResultRepository {
	return &JobResultRepository{db: db}
}

func (r *JobResultRepository) Upsert(ctx context.Context, tx *sqlx.Tx, result domain.JobResult) error {
	query := `
INSERT INTO dsl_job_results (job_id, result, result_bytes, created_at)
VALUES (:job_id, :result, :result_bytes, :created_at)
ON CONFLICT (job_id)
DO UPDATE SET result = EXCLUDED.result,
              result_bytes = EXCLUDED.result_bytes,
              created_at = EXCLUDED.created_at`
	if tx != nil {
		_, err := tx.NamedExecContext(ctx, query, result)
		return err
	}
	_, err := r.db.NamedExecContext(ctx, query, result)
	return err
}

func (r *JobResultRepository) GetByJobID(ctx context.Context, jobID string) (domain.JobResult, error) {
	var res domain.JobResult
	query := `
SELECT job_id, result, result_bytes, created_at
FROM dsl_job_results
WHERE job_id = $1`
	if err := r.db.GetContext(ctx, &res, query, jobID); err != nil {
		if err == sql.ErrNoRows {
			return domain.JobResult{}, ErrNotFound
		}
		return domain.JobResult{}, err
	}
	return res, nil
}
