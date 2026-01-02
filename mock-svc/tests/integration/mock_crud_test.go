package integration

import (
	"encoding/json"
	"net/http"
	"testing"
)

type mockResponse struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Description      *string         `json:"description"`
	RequestMatch     json.RawMessage `json:"requestMatch"`
	ResponseTemplate json.RawMessage `json:"responseTemplate"`
	Enabled          bool            `json:"enabled"`
	FamilyID         *string         `json:"familyId"`
	CreatedBy        string          `json:"createdBy"`
	GeneratedBy      *string         `json:"generatedBy"`
	CreatedAt        string          `json:"createdAt"`
	UpdatedAt        string          `json:"updatedAt"`
	DeletedAt        *string         `json:"deletedAt"`
}

func TestMockCRUDAndSearch(t *testing.T) {
	srv := newMockServer(t)
	defer srv.CleanupFn()

	client := srv.Server.Client()
	headers := map[string]string{
		"X-User-Id": "00000000-0000-0000-0000-000000000001",
	}

	createBody := map[string]any{
		"name": "test-mock",
		"requestMatch": map[string]any{
			"method": "GET",
			"path":   "/api/test",
		},
		"responseTemplate": map[string]any{
			"status": 200,
			"body": map[string]any{
				"value": "ok",
			},
		},
	}
	status, body := doJSON(t, client, http.MethodPost, srv.Server.URL+"/mocks/v1/mocks", createBody, headers)
	if status != http.StatusCreated {
		t.Fatalf("create mock expected 201, got %d: %s", status, string(body))
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if created.ID == "" {
		t.Fatalf("missing id in create response")
	}
	srv.Auth.AddResource("mock", created.ID)

	status, body = doJSON(t, client, http.MethodGet, srv.Server.URL+"/mocks/v1/mocks/"+created.ID, nil, headers)
	if status != http.StatusOK {
		t.Fatalf("get mock expected 200, got %d: %s", status, string(body))
	}
	var got mockResponse
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode get: %v", err)
	}
	if got.Name != "test-mock" || got.CreatedBy != "00000000-0000-0000-0000-000000000001" {
		t.Fatalf("unexpected mock data: %+v", got)
	}

	updateBody := map[string]any{
		"name":        "updated-mock",
		"enabled":     false,
		"description": "desc",
	}
	status, body = doJSON(t, client, http.MethodPatch, srv.Server.URL+"/mocks/v1/mocks/"+created.ID, updateBody, headers)
	if status != http.StatusOK {
		t.Fatalf("update mock expected 200, got %d: %s", status, string(body))
	}
	var updated mockResponse
	if err := json.Unmarshal(body, &updated); err != nil {
		t.Fatalf("decode update: %v", err)
	}
	if updated.Name != "updated-mock" || updated.Enabled != false || updated.Description == nil || *updated.Description != "desc" {
		t.Fatalf("update mismatch: %+v", updated)
	}

	status, body = doJSON(t, client, http.MethodGet, srv.Server.URL+"/mocks/v1/mocks?method=GET&path=/api/test", nil, headers)
	if status != http.StatusOK {
		t.Fatalf("list mocks expected 200, got %d: %s", status, string(body))
	}
	var listResp struct {
		Items []mockResponse `json:"items"`
	}
	if err := json.Unmarshal(body, &listResp); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listResp.Items) != 1 || listResp.Items[0].ID != created.ID {
		t.Fatalf("list result mismatch: %+v", listResp.Items)
	}

	status, body = doJSON(t, client, http.MethodDelete, srv.Server.URL+"/mocks/v1/mocks/"+created.ID, nil, headers)
	if status != http.StatusNoContent {
		t.Fatalf("delete mock expected 204, got %d: %s", status, string(body))
	}

	status, body = doJSON(t, client, http.MethodGet, srv.Server.URL+"/mocks/v1/mocks?method=GET&path=/api/test", nil, headers)
	if status != http.StatusOK {
		t.Fatalf("list after delete expected 200, got %d: %s", status, string(body))
	}
	if err := json.Unmarshal(body, &listResp); err != nil {
		t.Fatalf("decode list after delete: %v", err)
	}
	if len(listResp.Items) != 0 {
		t.Fatalf("expected empty list after delete, got %d", len(listResp.Items))
	}
}
