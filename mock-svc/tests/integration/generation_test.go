package integration

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"mock-svc/internal/application"
)

func TestGenerationApplyCreatesDerivedMocks(t *testing.T) {
	srv := newMockServer(t)
	defer srv.CleanupFn()

	client := srv.Server.Client()
	headers := map[string]string{
		"X-User-Id": "00000000-0000-0000-0000-000000000001",
	}

	// base mock
	createBody := map[string]any{
		"name": "base-mock",
		"requestMatch": map[string]any{
			"method": "GET",
			"path":   "/base",
		},
		"responseTemplate": map[string]any{
			"status": 200,
		},
	}
	status, body := doJSON(t, client, http.MethodPost, srv.Server.URL+"/mocks/v1/mocks", createBody, headers)
	if status != http.StatusCreated {
		t.Fatalf("create base mock expected 201, got %d: %s", status, string(body))
	}
	var created struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(body, &created)
	if created.ID == "" {
		t.Fatalf("missing base mock id")
	}
	srv.Auth.AddResource("mock", created.ID)
	srv.Auth.AddResource("dsl_script", "11111111-1111-1111-1111-111111111111")

	derivedMock := application.GeneratedMock{
		Name: "derived-1",
		RequestMatch: json.RawMessage(`{
			"method": "GET",
			"path": "/base/derived"
		}`),
		ResponseTemplate: json.RawMessage(`{"status":201}`),
	}
	srv.DSL.SetDerived([]application.GeneratedMock{derivedMock})

	genBody := map[string]any{
		"baseMockId":  created.ID,
		"dslScriptId": "11111111-1111-1111-1111-111111111111",
		"mode":        "apply",
	}
	status, body = doJSON(t, client, http.MethodPost, srv.Server.URL+"/mocks/v1/generations", genBody, headers)
	if status != http.StatusAccepted {
		t.Fatalf("start generation expected 202, got %d: %s", status, string(body))
	}
	var genResp struct {
		GenerationID string `json:"generationId"`
		Status       string `json:"status"`
	}
	_ = json.Unmarshal(body, &genResp)
	if genResp.GenerationID == "" {
		t.Fatalf("missing generation id")
	}

	// poll until DONE
	var final struct {
		GenerationID  string   `json:"generationId"`
		Status        string   `json:"status"`
		ResultMockIDs []string `json:"resultMockIds"`
		Error         *string  `json:"error"`
	}
	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)
		status, body = doJSON(t, client, http.MethodGet, srv.Server.URL+"/mocks/v1/generations/"+genResp.GenerationID, nil, headers)
		if status != http.StatusOK {
			t.Fatalf("get generation expected 200, got %d: %s", status, string(body))
		}
		_ = json.Unmarshal(body, &final)
		if final.Status == "DONE" || final.Status == "FAILED" {
			break
		}
	}
	if final.Status != "DONE" {
		t.Fatalf("generation not completed: %+v", final)
	}
	if len(final.ResultMockIDs) != 1 {
		t.Fatalf("expected 1 derived mock, got %d", len(final.ResultMockIDs))
	}

	// check derived mock exists
	derivedID := final.ResultMockIDs[0]
	srv.Auth.AddResource("mock", derivedID)
	status, body = doJSON(t, client, http.MethodGet, srv.Server.URL+"/mocks/v1/mocks/"+derivedID, nil, headers)
	if status != http.StatusOK {
		t.Fatalf("get derived mock expected 200, got %d: %s", status, string(body))
	}
	var got mockResponse
	_ = json.Unmarshal(body, &got)
	if got.Name != "derived-1" || got.GeneratedBy == nil || *got.GeneratedBy != "00000000-0000-0000-0000-000000000001" {
		t.Fatalf("unexpected derived mock: %+v", got)
	}
}
