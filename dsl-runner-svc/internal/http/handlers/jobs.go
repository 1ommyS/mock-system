package handlers

import (
	"net/http"
	"strings"
	"time"

	"dsl-runner-svc/internal/domain"
)

type jobStatusResponse struct {
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

func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	jobID := strings.TrimPrefix(r.URL.Path, "/dslrunner/v1/jobs/")
	jobID = strings.TrimSuffix(jobID, "/result")
	jobID = strings.TrimSuffix(jobID, "/")

	job, err := h.Service.GetJobStatus(r.Context(), jobID)
	if err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}

	resp := jobStatusResponse{
		JobID:      job.JobID,
		Status:     job.Status,
		Attempt:    job.Attempt,
		CreatedAt:  job.CreatedAt.UTC().Format(time.RFC3339),
		ScriptHash: job.ScriptHash,
		InputHash:  job.InputHash,
	}
	if job.StartedAt != nil {
		val := job.StartedAt.UTC().Format(time.RFC3339)
		resp.StartedAt = &val
	}
	if job.FinishedAt != nil {
		val := job.FinishedAt.UTC().Format(time.RFC3339)
		resp.FinishedAt = &val
	}
	if job.ErrorCode != nil {
		resp.Error = &domain.JobError{Code: *job.ErrorCode}
		if job.ErrorMessage != nil {
			resp.Error.Message = *job.ErrorMessage
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetJobResult(w http.ResponseWriter, r *http.Request) {
	jobID := strings.TrimPrefix(r.URL.Path, "/dslrunner/v1/jobs/")
	jobID = strings.TrimSuffix(jobID, "/result")
	jobID = strings.TrimSuffix(jobID, "/")

	result, err := h.Service.GetJobResult(r.Context(), jobID)
	if err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result)
}
