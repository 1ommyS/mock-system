package application

import (
	"context"
	"time"

	"dsl-runner-svc/internal/domain"
	"dsl-runner-svc/internal/dsl"

	"github.com/jmoiron/sqlx"
)

type TxManager interface {
	WithTx(ctx context.Context, fn func(*sqlx.Tx) error) error
}

type JobRepository interface {
	CreateIfAbsent(ctx context.Context, tx *sqlx.Tx, job domain.Job) (domain.Job, bool, error)
	GetByID(ctx context.Context, jobID string) (domain.Job, error)
	GetByKey(ctx context.Context, jobKey string) (domain.Job, error)
	UpdateStatus(ctx context.Context, tx *sqlx.Tx, jobID string, status string, attempt int, startedAt, finishedAt *time.Time, errorCode, errorMessage *string) error
}

type JobResultRepository interface {
	Upsert(ctx context.Context, tx *sqlx.Tx, result domain.JobResult) error
	GetByJobID(ctx context.Context, jobID string) (domain.JobResult, error)
}

type Repositories struct {
	Jobs    JobRepository
	Results JobResultRepository
}

type ResultPublisher interface {
	Publish(ctx context.Context, msg domain.ResultMessage) error
}

type Service struct {
	Repos       Repositories
	Tx          TxManager
	Publisher   ResultPublisher
	Executor    Executor
	MaxAttempts int
}

type Executor interface {
	Run(input dsl.ExecInput) (dsl.ExecResult, error)
}

func New(repos Repositories, tx TxManager, publisher ResultPublisher, executor Executor, maxAttempts int) *Service {
	return &Service{Repos: repos, Tx: tx, Publisher: publisher, Executor: executor, MaxAttempts: maxAttempts}
}
