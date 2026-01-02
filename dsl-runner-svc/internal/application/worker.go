package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"dsl-runner-svc/internal/domain"
	"dsl-runner-svc/internal/dsl"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type JobHandler struct {
	Service *Service
}

func (h *JobHandler) Process(ctx context.Context, msg domain.JobMessage) error {
	if err := validateJobMessage(msg); err != nil {
		return err
	}
	if msg.DSLScript == nil || *msg.DSLScript == "" {
		return NewCodeError(ErrCodeInvalidJobPayload, fmt.Errorf("dslScript required"))
	}

	scriptText := *msg.DSLScript

	scriptHash := dsl.Sha256Hex([]byte(scriptText))
	baseCanon, err := dsl.CanonicalJSON(msg.BaseMock)
	if err != nil {
		return NewCodeError(ErrCodeDSLValidation, err)
	}
	paramsCanon, err := dsl.CanonicalJSON(msg.Params)
	if err != nil {
		return NewCodeError(ErrCodeDSLValidation, err)
	}
	inputHash := dsl.Sha256Hex([]byte(baseCanon + paramsCanon + msg.Seed + scriptHash))

	createdAt := msg.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	job := domain.Job{
		JobID:       msg.JobID,
		JobKey:      msg.JobKey,
		Status:      domain.JobStatusQueued,
		Mode:        msg.Mode,
		BaseMockID:  msg.BaseMockID,
		DSLScriptID: msg.DSLScriptID,
		ScriptHash:  scriptHash,
		InputHash:   inputHash,
		Seed:        msg.Seed,
		Params:      msg.Params,
		Limits:      mustJSON(msg.Limits),
		Attempt:     0,
		CreatedAt:   createdAt,
	}

	created, inserted, err := h.Service.Repos.Jobs.CreateIfAbsent(ctx, nil, job)
	if err != nil {
		return err
	}
	job = created

	if !inserted {
		switch job.Status {
		case domain.JobStatusDone, domain.JobStatusFailed:
			return h.publishResult(ctx, job)
		case domain.JobStatusRunning:
			return NewTransient(fmt.Errorf("job %s already running", job.JobID))
		}
	}

	attempt := job.Attempt + 1
	if attempt > h.Service.MaxAttempts {
		err := fmt.Errorf("max attempts exceeded")
		h.handleFailure(ctx, job.JobID, attempt, err)
		return nil
	}
	startedAt := time.Now().UTC()
	if err := h.Service.Repos.Jobs.UpdateStatus(ctx, nil, job.JobID, domain.JobStatusRunning, attempt, &startedAt, nil, nil, nil); err != nil {
		return err
	}

	execInput := dsl.ExecInput{
		Seed:       msg.Seed,
		ScriptText: scriptText,
		BaseMock:   msg.BaseMock,
		Params:     msg.Params,
		Limits:     msg.Limits,
	}

	execRes, err := h.Service.Executor.Run(execInput)
	if err != nil {
		h.handleFailure(ctx, job.JobID, attempt, err)
		return err
	}

	payload := domain.JobResultPayload{
		JobID:      job.JobID,
		ScriptHash: execRes.ScriptHash,
		InputHash:  execRes.InputHash,
		Generated:  execRes.Generated,
		Warnings:   execRes.Warnings,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		h.handleFailure(ctx, job.JobID, attempt, err)
		return err
	}
	res := domain.JobResult{
		JobID:       job.JobID,
		Result:      payloadBytes,
		ResultBytes: len(payloadBytes),
		CreatedAt:   time.Now().UTC(),
	}
	finishedAt := res.CreatedAt
	if err := h.Service.Tx.WithTx(ctx, func(tx *sqlx.Tx) error {
		if err := h.Service.Repos.Results.Upsert(ctx, tx, res); err != nil {
			return err
		}
		return h.Service.Repos.Jobs.UpdateStatus(ctx, tx, job.JobID, domain.JobStatusDone, attempt, &startedAt, &finishedAt, nil, nil)
	}); err != nil {
		return err
	}

	return h.publishResult(ctx, domain.Job{
		JobID:      job.JobID,
		JobKey:     job.JobKey,
		Status:     domain.JobStatusDone,
		ScriptHash: execRes.ScriptHash,
		InputHash:  execRes.InputHash,
		Attempt:    attempt,
		FinishedAt: &finishedAt,
	})
}

func (h *JobHandler) handleFailure(ctx context.Context, jobID string, attempt int, err error) {
	code := CodeFromError(err)
	msg := err.Error()
	finished := time.Now().UTC()
	_ = h.Service.Repos.Jobs.UpdateStatus(ctx, nil, jobID, domain.JobStatusFailed, attempt, nil, &finished, &code, &msg)
}

func (h *JobHandler) publishResult(ctx context.Context, job domain.Job) error {
	var generatedCount int
	var resultBytes int
	if job.Status == domain.JobStatusDone {
		if res, err := h.Service.Repos.Results.GetByJobID(ctx, job.JobID); err == nil {
			resultBytes = res.ResultBytes
			var payload domain.JobResultPayload
			_ = json.Unmarshal(res.Result, &payload)
			generatedCount = len(payload.Generated)
		}
	}

	msg := domain.ResultMessage{
		JobID:          job.JobID,
		JobKey:         job.JobKey,
		Status:         job.Status,
		ScriptHash:     job.ScriptHash,
		InputHash:      job.InputHash,
		GeneratedCount: generatedCount,
		ResultBytes:    resultBytes,
		FinishedAt:     time.Now().UTC(),
	}
	if job.FinishedAt != nil {
		msg.FinishedAt = *job.FinishedAt
	}
	if job.Status == domain.JobStatusFailed {
		code := "UNKNOWN"
		if job.ErrorCode != nil {
			code = *job.ErrorCode
		}
		message := "failed"
		if job.ErrorMessage != nil && *job.ErrorMessage != "" {
			message = *job.ErrorMessage
		}
		msg.Error = &domain.JobError{Code: code, Message: message}
	}
	return h.Service.Publisher.Publish(ctx, msg)
}

func validateJobMessage(msg domain.JobMessage) error {
	if msg.JobID == "" || msg.JobKey == "" {
		return NewCodeError(ErrCodeInvalidJobPayload, fmt.Errorf("jobId and jobKey required"))
	}
	if _, err := uuid.Parse(msg.JobID); err != nil {
		return NewCodeError(ErrCodeInvalidJobPayload, fmt.Errorf("jobId invalid"))
	}
	if msg.Mode != domain.JobModePreview && msg.Mode != domain.JobModeApply {
		return NewCodeError(ErrCodeInvalidJobPayload, fmt.Errorf("mode invalid"))
	}
	if msg.Limits.TimeoutMs <= 0 || msg.Limits.MaxGeneratedMocks <= 0 || msg.Limits.MaxResultBytes <= 0 {
		return NewCodeError(ErrCodeInvalidJobPayload, fmt.Errorf("limits invalid"))
	}
	return nil
}

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
