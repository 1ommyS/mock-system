package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"mock-svc/internal/application"
)

type DSLRunnerClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewDSLRunnerClient(baseURL string) *DSLRunnerClient {
	return &DSLRunnerClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *DSLRunnerClient) Preview(ctx context.Context, baseMock application.DSLBaseMock, dslScriptID string, params json.RawMessage) ([]application.GeneratedMock, error) {
	payload := map[string]any{
		"baseMock":    baseMock,
		"dslScriptId": dslScriptID,
		"params":      params,
	}
	resp, status, err := c.doJSON(ctx, http.MethodPost, "/dsl-runner/v1/preview", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("dsl preview failed: %s", resp.Status)
	}
	var out struct {
		Planned []application.GeneratedMock `json:"planned"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Planned, nil
}

func (c *DSLRunnerClient) Apply(ctx context.Context, baseMock application.DSLBaseMock, dslScriptID string, params json.RawMessage) ([]application.GeneratedMock, error) {
	payload := map[string]any{
		"baseMock":    baseMock,
		"dslScriptId": dslScriptID,
		"params":      params,
	}
	resp, status, err := c.doJSON(ctx, http.MethodPost, "/dsl-runner/v1/apply", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("dsl apply failed: %s", resp.Status)
	}
	var out struct {
		Derived []application.GeneratedMock `json:"derived"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Derived, nil
}

func (c *DSLRunnerClient) doJSON(ctx context.Context, method, path string, payload any) (*http.Response, int, error) {
	endpoint := c.baseURL + path
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	return resp, resp.StatusCode, nil
}
