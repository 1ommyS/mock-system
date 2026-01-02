package domain

import (
	"encoding/json"
	"time"
)

const (
	JobStatusQueued  = "QUEUED"
	JobStatusRunning = "RUNNING"
	JobStatusDone    = "DONE"
	JobStatusFailed  = "FAILED"

	JobModePreview = "preview"
	JobModeApply   = "apply"
)

type JobLimits struct {
	TimeoutMs         int `json:"timeoutMs"`
	MaxGeneratedMocks int `json:"maxGeneratedMocks"`
	MaxResultBytes    int `json:"maxResultBytes"`
}

type JobMessage struct {
	JobID       string          `json:"jobId"`
	JobKey      string          `json:"jobKey"`
	CreatedAt   time.Time       `json:"createdAt"`
	Mode        string          `json:"mode"`
	Seed        string          `json:"seed"`
	BaseMockID  *string         `json:"baseMockId,omitempty"`
	DSLScriptID *string         `json:"dslScriptId,omitempty"`
	DSLScript   *string         `json:"dslScript,omitempty"`
	BaseMock    json.RawMessage `json:"baseMock"`
	Params      json.RawMessage `json:"params"`
	Limits      JobLimits       `json:"limits"`
}

type Job struct {
	JobID        string          `db:"job_id"`
	JobKey       string          `db:"job_key"`
	Status       string          `db:"status"`
	Mode         string          `db:"mode"`
	BaseMockID   *string         `db:"base_mock_id"`
	DSLScriptID  *string         `db:"dsl_script_id"`
	ScriptHash   string          `db:"script_hash"`
	InputHash    string          `db:"input_hash"`
	Seed         string          `db:"seed"`
	Params       json.RawMessage `db:"params"`
	Limits       json.RawMessage `db:"limits"`
	Attempt      int             `db:"attempt"`
	ErrorCode    *string         `db:"error_code"`
	ErrorMessage *string         `db:"error_message"`
	CreatedAt    time.Time       `db:"created_at"`
	StartedAt    *time.Time      `db:"started_at"`
	FinishedAt   *time.Time      `db:"finished_at"`
}

type JobResult struct {
	JobID       string          `db:"job_id"`
	Result      json.RawMessage `db:"result"`
	ResultBytes int             `db:"result_bytes"`
	CreatedAt   time.Time       `db:"created_at"`
}

type JobError struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}

type ResultMessage struct {
	JobID          string    `json:"jobId"`
	JobKey         string    `json:"jobKey"`
	Status         string    `json:"status"`
	ScriptHash     string    `json:"scriptHash"`
	InputHash      string    `json:"inputHash"`
	GeneratedCount int       `json:"generatedCount"`
	ResultBytes    int       `json:"resultBytes"`
	Error          *JobError `json:"error,omitempty"`
	FinishedAt     time.Time `json:"finishedAt"`
}

type GeneratedMockSpec struct {
	Name             string          `json:"name"`
	RequestMatch     json.RawMessage `json:"requestMatch"`
	ResponseTemplate json.RawMessage `json:"responseTemplate"`
	Meta             json.RawMessage `json:"meta"`
	DerivedSeed      string          `json:"derivedSeed"`
}

type JobResultPayload struct {
	JobID      string              `json:"jobId"`
	ScriptHash string              `json:"scriptHash"`
	InputHash  string              `json:"inputHash"`
	Generated  []GeneratedMockSpec `json:"generated"`
	Warnings   []string            `json:"warnings"`
}
