package application

import (
	"context"
	"encoding/json"

	"dsl-runner-svc/internal/domain"

	"github.com/google/uuid"
)

type JobStatus struct {
	JobID      string           `json:"jobId"`
	Status     string           `json:"status"`
	Attempt    int              `json:"attempt"`
	CreatedAt  string           `json:"createdAt"`
	StartedAt  *string          `json:"startedAt,omitempty"`
	FinishedAt *string          `json:"finishedAt,omitempty"`
	ScriptHash string           `json:"scriptHash"`
	InputHash  string           `json:"inputHash"`
	Error      *domain.JobError `json:"error,omitempty"`
}

func (s *Service) GetJobStatus(ctx context.Context, jobID string) (domain.Job, error) {
	if jobID == "" {
		return domain.Job{}, ErrInvalidRequest
	}
	if _, err := uuid.Parse(jobID); err != nil {
		return domain.Job{}, ErrInvalidRequest
	}
	job, err := s.Repos.Jobs.GetByID(ctx, jobID)
	if err != nil {
		return domain.Job{}, mapRepoError(err)
	}
	return job, nil
}

func (s *Service) GetJobResult(ctx context.Context, jobID string) (json.RawMessage, error) {
	if jobID == "" {
		return nil, ErrInvalidRequest
	}
	if _, err := uuid.Parse(jobID); err != nil {
		return nil, ErrInvalidRequest
	}
	job, err := s.Repos.Jobs.GetByID(ctx, jobID)
	if err != nil {
		return nil, mapRepoError(err)
	}
	if job.Status != domain.JobStatusDone {
		return nil, ErrJobNotReady
	}
	res, err := s.Repos.Results.GetByJobID(ctx, jobID)
	if err != nil {
		return nil, mapRepoError(err)
	}
	return res.Result, nil
}

func mapRepoError(err error) error {
	if err == nil {
		return nil
	}
	if err == domain.ErrNotFound {
		return ErrNotFound
	}
	return err
}
