package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"mock-svc/internal/application"
)

type JobPublisher interface {
	Publish(ctx context.Context, key, value []byte) error
}

type DSLRunnerClient struct {
	baseURL        string
	httpClient     *http.Client
	publisher      JobPublisher
	internalSecret string
	headerName     string
	pollInterval   time.Duration
	pollTimeout    time.Duration
	limits         JobLimits
	jobIDFn        func() string
	nowFn          func() time.Time
}

type JobLimits struct {
	TimeoutMs         int `json:"timeoutMs"`
	MaxGeneratedMocks int `json:"maxGeneratedMocks"`
	MaxResultBytes    int `json:"maxResultBytes"`
}

type jobMessage struct {
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

type jobResult struct {
	JobID      string          `json:"jobId"`
	ScriptHash string          `json:"scriptHash"`
	InputHash  string          `json:"inputHash"`
	Generated  []generatedSpec `json:"generated"`
	Warnings   []string        `json:"warnings"`
}

type jobStatus struct {
	Status string `json:"status"`
	Error  *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type generatedSpec struct {
	Name             string          `json:"name"`
	RequestMatch     json.RawMessage `json:"requestMatch"`
	ResponseTemplate json.RawMessage `json:"responseTemplate"`
	Meta             json.RawMessage `json:"meta"`
	DerivedSeed      string          `json:"derivedSeed"`
}

type DSLRunnerConfig struct {
	BaseURL        string
	InternalSecret string
	HeaderName     string
	PollInterval   time.Duration
	PollTimeout    time.Duration
	JobsTopic      string
	Limits         JobLimits
	Publisher      JobPublisher
	JobIDFn        func() string
	NowFn          func() time.Time
}

func NewDSLRunnerClient(cfg DSLRunnerConfig) *DSLRunnerClient {
	header := cfg.HeaderName
	if header == "" {
		header = "X-Internal-Secret"
	}
	pollInterval := cfg.PollInterval
	if pollInterval == 0 {
		pollInterval = 200 * time.Millisecond
	}
	pollTimeout := cfg.PollTimeout
	if pollTimeout == 0 {
		pollTimeout = 5 * time.Second
	}
	jobIDFn := cfg.JobIDFn
	nowFn := cfg.NowFn
	if nowFn == nil {
		nowFn = func() time.Time { return time.Now().UTC() }
	}

	return &DSLRunnerClient{
		baseURL:        strings.TrimRight(cfg.BaseURL, "/"),
		httpClient:     &http.Client{Timeout: 10 * time.Second},
		publisher:      cfg.Publisher,
		internalSecret: cfg.InternalSecret,
		headerName:     header,
		pollInterval:   pollInterval,
		pollTimeout:    pollTimeout,
		limits:         cfg.Limits,
		jobIDFn:        jobIDFn,
		nowFn:          nowFn,
	}
}

func (c *DSLRunnerClient) Preview(ctx context.Context, baseMock application.DSLBaseMock, dslScriptID string, dslScript string, params json.RawMessage) ([]application.GeneratedMock, error) {
	return c.run(ctx, "preview", baseMock, dslScriptID, dslScript, params)
}

func (c *DSLRunnerClient) Apply(ctx context.Context, baseMock application.DSLBaseMock, dslScriptID string, dslScript string, params json.RawMessage) ([]application.GeneratedMock, error) {
	return c.run(ctx, "apply", baseMock, dslScriptID, dslScript, params)
}

func (c *DSLRunnerClient) run(ctx context.Context, mode string, baseMock application.DSLBaseMock, dslScriptID string, dslScript string, params json.RawMessage) ([]application.GeneratedMock, error) {
	if dslScript == "" {
		return nil, fmt.Errorf("dslScript is required")
	}
	createdAt := c.nowFn()
	baseMockPayload, err := buildBaseMockPayload(baseMock)
	if err != nil {
		return nil, err
	}
	params = ensureJSON(params)

	baseCanon, err := canonicalJSON(baseMockPayload)
	if err != nil {
		return nil, err
	}
	paramsCanon, err := canonicalJSON(params)
	if err != nil {
		return nil, err
	}

	scriptHash := sha256Hex([]byte(dslScript))
	seed := sha256Hex([]byte(scriptHash + ":" + baseCanon + ":" + paramsCanon + ":" + mode))
	inputHash := sha256Hex([]byte(baseCanon + paramsCanon + seed + scriptHash))
	jobKey := sha256Hex([]byte(scriptHash + ":" + inputHash + ":" + mode))
	jobID := ""
	if c.jobIDFn != nil {
		jobID = c.jobIDFn()
	}
	if jobID == "" {
		jobID = deterministicUUID(jobKey)
	}

	slog.Info(
		"dsl runner job enqueue",
		"job_id",
		jobID,
		"job_key",
		jobKey,
		"mode",
		mode,
		"base_mock_id",
		baseMock.ID,
		"dsl_script_id",
		dslScriptID,
		"script_hash",
		scriptHash,
	)

	msg := jobMessage{
		JobID:       jobID,
		JobKey:      jobKey,
		CreatedAt:   createdAt,
		Mode:        mode,
		Seed:        seed,
		BaseMockID:  optionalString(baseMock.ID),
		DSLScriptID: optionalString(dslScriptID),
		DSLScript:   &dslScript,
		BaseMock:    baseMockPayload,
		Params:      params,
		Limits:      c.limits,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	if err := c.publisher.Publish(ctx, []byte(jobKey), payload); err != nil {
		slog.Error(
			"dsl runner publish failed",
			"job_id",
			jobID,
			"job_key",
			jobKey,
			"mode",
			mode,
			"error",
			err,
		)
		return nil, err
	}

	pollCtx := ctx
	if c.pollTimeout > 0 {
		var cancel context.CancelFunc
		pollCtx, cancel = context.WithTimeout(ctx, c.pollTimeout)
		defer cancel()
	}
	result, err := c.waitForResult(pollCtx, jobID)
	if err != nil {
		slog.Error(
			"dsl runner result failed",
			"job_id",
			jobID,
			"job_key",
			jobKey,
			"mode",
			mode,
			"error",
			err,
		)
		return nil, err
	}
	slog.Info(
		"dsl runner result ready",
		"job_id",
		jobID,
		"job_key",
		jobKey,
		"mode",
		mode,
		"generated",
		len(result),
	)
	return result, nil
}

func (c *DSLRunnerClient) waitForResult(ctx context.Context, jobID string) ([]application.GeneratedMock, error) {
	endpoint := c.baseURL + "/dslrunner/v1/jobs/" + jobID + "/result"

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set(c.headerName, c.internalSecret)
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		data, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusOK:
			var result jobResult
			if err := json.Unmarshal(data, &result); err != nil {
				return nil, err
			}
			return mapGenerated(result.Generated), nil
		case http.StatusConflict, http.StatusNotFound:
			status, err := c.getJobStatus(ctx, jobID)
			if err != nil {
				return nil, err
			}
			if status != nil && status.Status == "FAILED" {
				if status.Error != nil && status.Error.Message != "" {
					return nil, fmt.Errorf("dsl runner job failed (%s): %s", status.Error.Code, status.Error.Message)
				}
				if status.Error != nil && status.Error.Code != "" {
					return nil, fmt.Errorf("dsl runner job failed (%s)", status.Error.Code)
				}
				return nil, fmt.Errorf("dsl runner job failed")
			}
			select {
			case <-time.After(c.pollInterval):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		default:
			return nil, fmt.Errorf("dsl runner error: %s", string(data))
		}
	}
}

func (c *DSLRunnerClient) getJobStatus(ctx context.Context, jobID string) (*jobStatus, error) {
	endpoint := c.baseURL + "/dslrunner/v1/jobs/" + jobID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(c.headerName, c.internalSecret)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	data, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusConflict {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dsl runner status error: %s", string(data))
	}

	var status jobStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

func mapGenerated(items []generatedSpec) []application.GeneratedMock {
	out := make([]application.GeneratedMock, 0, len(items))
	for _, item := range items {
		out = append(out, application.GeneratedMock{
			Name:             item.Name,
			RequestMatch:     item.RequestMatch,
			ResponseTemplate: item.ResponseTemplate,
		})
	}
	return out
}

func buildBaseMockPayload(base application.DSLBaseMock) (json.RawMessage, error) {
	payload := map[string]any{
		"id":                base.ID,
		"name":              base.Name,
		"description":       base.Description,
		"request_match":     json.RawMessage(base.RequestMatch),
		"requestMatch":      json.RawMessage(base.RequestMatch),
		"response_template": json.RawMessage(base.ResponseTemplate),
		"responseTemplate":  json.RawMessage(base.ResponseTemplate),
		"enabled":           base.Enabled,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ensureJSON(data json.RawMessage) json.RawMessage {
	if len(data) == 0 {
		return json.RawMessage(`{}`)
	}
	return data
}

func canonicalJSON(data json.RawMessage) (string, error) {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return "", err
	}
	canon, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(canon), nil
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func optionalString(val string) *string {
	if val == "" {
		return nil
	}
	return &val
}

func newUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func deterministicUUID(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	b := sum[:16]
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
