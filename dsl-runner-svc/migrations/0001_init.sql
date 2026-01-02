-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS dsl_jobs
(
    job_id       uuid PRIMARY KEY,
    job_key      text        NOT NULL UNIQUE,
    status       text        NOT NULL CHECK (status IN ('QUEUED', 'RUNNING', 'DONE', 'FAILED')),
    mode         text        NOT NULL CHECK (mode IN ('preview', 'apply')),
    base_mock_id uuid        NULL,
    dsl_script_id uuid       NULL,
    script_hash  text        NOT NULL,
    input_hash   text        NOT NULL,
    seed         text        NOT NULL,
    params       jsonb       NULL,
    limits       jsonb       NULL,
    attempt      int         NOT NULL DEFAULT 0,
    error_code   text        NULL,
    error_message text       NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    started_at   timestamptz NULL,
    finished_at  timestamptz NULL
);

CREATE TABLE IF NOT EXISTS dsl_job_results
(
    job_id      uuid PRIMARY KEY REFERENCES dsl_jobs (job_id) ON DELETE CASCADE,
    result      jsonb       NOT NULL,
    result_bytes int        NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_dsl_jobs_status_created ON dsl_jobs (status, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS dsl_job_results;
DROP TABLE IF EXISTS dsl_jobs;
-- +goose StatementEnd
