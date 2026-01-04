package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"mock-svc/internal/application"
	"mock-svc/internal/service"
)

type dslRunnerInput struct {
	JobID       string          `json:"jobId"`
	CreatedAt   string          `json:"createdAt"`
	BaseMock    inputBaseMock   `json:"baseMock"`
	DSLScriptID string          `json:"dslScriptId"`
	DSLScript   string          `json:"dslScript"`
	Params      json.RawMessage `json:"params"`
}

type inputBaseMock struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	RequestMatch     json.RawMessage `json:"requestMatch"`
	ResponseTemplate json.RawMessage `json:"responseTemplate"`
	Enabled          bool            `json:"enabled"`
}

type jobMessage struct {
	JobKey string `json:"jobKey"`
}

type dslResult struct {
	Generated []struct {
		Name             string          `json:"name"`
		RequestMatch     json.RawMessage `json:"requestMatch"`
		ResponseTemplate json.RawMessage `json:"responseTemplate"`
	} `json:"generated"`
}

type capturePublisher struct {
	key   []byte
	value []byte
}

func (p *capturePublisher) Publish(_ context.Context, key, value []byte) error {
	p.key = append([]byte(nil), key...)
	p.value = append([]byte(nil), value...)
	return nil
}

func TestDSLRunnerKafkaClient(t *testing.T) {
	inputBytes, err := os.ReadFile("data/dsl_runner_input.json")
	if err != nil {
		t.Fatalf("read input: %v", err)
	}
	var input dslRunnerInput
	if err := json.Unmarshal(inputBytes, &input); err != nil {
		t.Fatalf("parse input: %v", err)
	}

	jobExpected, err := os.ReadFile("data/dsl_runner_job_expected.json")
	if err != nil {
		t.Fatalf("read expected job: %v", err)
	}
	resultBytes, err := os.ReadFile("data/dsl_runner_result.json")
	if err != nil {
		t.Fatalf("read result: %v", err)
	}

	secret := "test-secret"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Internal-Secret") != secret {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/result") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(resultBytes)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	pub := &capturePublisher{}
	createdAt, err := time.Parse(time.RFC3339, input.CreatedAt)
	if err != nil {
		t.Fatalf("parse createdAt: %v", err)
	}

	client := service.NewDSLRunnerClient(service.DSLRunnerConfig{
		BaseURL:        srv.URL,
		InternalSecret: secret,
		PollInterval:   10 * time.Millisecond,
		PollTimeout:    500 * time.Millisecond,
		Publisher:      pub,
		Limits: service.JobLimits{
			TimeoutMs:         5000,
			MaxGeneratedMocks: 200,
			MaxResultBytes:    2000000,
		},
		JobIDFn: func() string { return input.JobID },
		NowFn:   func() time.Time { return createdAt },
	})

	base := application.DSLBaseMock{
		ID:               input.BaseMock.ID,
		Name:             input.BaseMock.Name,
		Description:      &input.BaseMock.Description,
		RequestMatch:     input.BaseMock.RequestMatch,
		ResponseTemplate: input.BaseMock.ResponseTemplate,
		Enabled:          input.BaseMock.Enabled,
	}

	mocks, err := client.Apply(context.Background(), base, input.DSLScriptID, input.DSLScript, input.Params)
	if err != nil {
		t.Fatalf("apply error: %v", err)
	}

	actualJobCanon := canonicalJSON(pub.value)
	expectedJobCanon := canonicalJSON(jobExpected)
	if actualJobCanon != expectedJobCanon {
		t.Fatalf("job payload mismatch\nactual: %s\nexpected: %s", actualJobCanon, expectedJobCanon)
	}

	var expectedJob jobMessage
	if err := json.Unmarshal(jobExpected, &expectedJob); err != nil {
		t.Fatalf("parse expected job: %v", err)
	}
	if string(pub.key) != expectedJob.JobKey {
		t.Fatalf("job key mismatch: %s vs %s", string(pub.key), expectedJob.JobKey)
	}

	var expectedResult dslResult
	if err := json.Unmarshal(resultBytes, &expectedResult); err != nil {
		t.Fatalf("parse result: %v", err)
	}
	if len(mocks) != len(expectedResult.Generated) {
		t.Fatalf("expected %d mocks, got %d", len(expectedResult.Generated), len(mocks))
	}
	for i, item := range expectedResult.Generated {
		if mocks[i].Name != item.Name {
			t.Fatalf("mock name mismatch: %s vs %s", mocks[i].Name, item.Name)
		}
		if canonicalJSON(mocks[i].RequestMatch) != canonicalJSON(item.RequestMatch) {
			t.Fatalf("requestMatch mismatch")
		}
		if canonicalJSON(mocks[i].ResponseTemplate) != canonicalJSON(item.ResponseTemplate) {
			t.Fatalf("responseTemplate mismatch")
		}
	}
}

func canonicalJSON(data []byte) string {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return string(data)
	}
	canon, err := json.Marshal(v)
	if err != nil {
		return string(data)
	}
	return string(canon)
}
