package postgres

import (
	"context"
	"database/sql"
	"time"

	"dsl-runner-svc/internal/domain"

	"github.com/jmoiron/sqlx"
)

type JobRepository struct {
	db *sqlx.DB
}

func NewJobRepository(db *sqlx.DB) *JobRepository {
	return &JobRepository{db: db}
}

func (r *JobRepository) CreateIfAbsent(ctx context.Context, tx *sqlx.Tx, job domain.Job) (domain.Job, bool, error) {
	query := `
INSERT INTO dsl_jobs
(job_id, job_key, status, mode, base_mock_id, dsl_script_id, script_hash, input_hash, seed, params, limits, attempt, error_code, error_message, created_at, started_at, finished_at)
VALUES
(:job_id, :job_key, :status, :mode, :base_mock_id, :dsl_script_id, :script_hash, :input_hash, :seed, :params, :limits, :attempt, :error_code, :error_message, :created_at, :started_at, :finished_at)`

	if tx != nil {
		_, err := tx.NamedExecContext(ctx, query, job)
		if err == nil {
			return job, true, nil
		}
		if IsUniqueViolation(err) {
			existing, err := r.GetByKey(ctx, job.JobKey)
			return existing, false, err
		}
		return domain.Job{}, false, err
	}

	_, err := r.db.NamedExecContext(ctx, query, job)
	if err == nil {
		return job, true, nil
	}
	if IsUniqueViolation(err) {
		existing, err := r.GetByKey(ctx, job.JobKey)
		return existing, false, err
	}
	return domain.Job{}, false, err
}

func (r *JobRepository) GetByID(ctx context.Context, jobID string) (domain.Job, error) {
	var job domain.Job
	query := `
SELECT job_id, job_key, status, mode, base_mock_id, dsl_script_id, script_hash, input_hash, seed, params, limits, attempt, error_code, error_message, created_at, started_at, finished_at
FROM dsl_jobs
WHERE job_id = $1`
	if err := r.db.GetContext(ctx, &job, query, jobID); err != nil {
		return domain.Job{}, mapNotFound(err)
	}
	return job, nil
}

func (r *JobRepository) GetByKey(ctx context.Context, jobKey string) (domain.Job, error) {
	var job domain.Job
	query := `
SELECT job_id, job_key, status, mode, base_mock_id, dsl_script_id, script_hash, input_hash, seed, params, limits, attempt, error_code, error_message, created_at, started_at, finished_at
FROM dsl_jobs
WHERE job_key = $1`
	if err := r.db.GetContext(ctx, &job, query, jobKey); err != nil {
		return domain.Job{}, mapNotFound(err)
	}
	return job, nil
}

func (r *JobRepository) UpdateStatus(ctx context.Context, tx *sqlx.Tx, jobID string, status string, attempt int, startedAt, finishedAt *time.Time, errorCode, errorMessage *string) error {
	query := `
UPDATE dsl_jobs
SET status = $2,
    attempt = $3,
    started_at = $4,
    finished_at = $5,
    error_code = $6,
    error_message = $7
WHERE job_id = $1`
	if tx != nil {
		_, err := tx.ExecContext(ctx, query, jobID, status, attempt, startedAt, finishedAt, errorCode, errorMessage)
		return err
	}
	_, err := r.db.ExecContext(ctx, query, jobID, status, attempt, startedAt, finishedAt, errorCode, errorMessage)
	return err
}

func mapNotFound(err error) error {
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	return err
}
