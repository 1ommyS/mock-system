package integration

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestFamilyCreateAddGet(t *testing.T) {
	srv := newMockServer(t)
	defer srv.CleanupFn()

	client := srv.Server.Client()
	headers := map[string]string{
		"X-User-Id": "00000000-0000-0000-0000-000000000001",
		"X-Roles":   "ADMIN",
	}

	createMock := func(name string) string {
		body := map[string]any{
			"name": name,
			"requestMatch": map[string]any{
				"method": "GET",
				"path":   "/" + name,
			},
			"responseTemplate": map[string]any{
				"status": 200,
			},
		}
		status, resp := doJSON(t, client, http.MethodPost, srv.Server.URL+"/mocks/v1/mocks", body, headers)
		if status != http.StatusCreated {
			t.Fatalf("create mock %s expected 201, got %d: %s", name, status, string(resp))
		}
		var out struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(resp, &out)
		if out.ID == "" {
			t.Fatalf("missing id for mock %s", name)
		}
		srv.Auth.AddResource("mock", out.ID)
		return out.ID
	}

	primaryID := createMock("primary")
	otherID := createMock("secondary")

	status, resp := doJSON(t, client, http.MethodPost, srv.Server.URL+"/mocks/v1/families", map[string]any{"primaryMockId": primaryID}, headers)
	if status != http.StatusCreated {
		t.Fatalf("create family expected 201, got %d: %s", status, string(resp))
	}
	var family struct {
		FamilyID string `json:"familyId"`
	}
	_ = json.Unmarshal(resp, &family)
	if family.FamilyID == "" {
		t.Fatalf("family id empty")
	}

	status, resp = doJSON(t, client, http.MethodPost, srv.Server.URL+"/mocks/v1/families/"+family.FamilyID+"/mocks", map[string]any{"mockId": otherID}, headers)
	if status != http.StatusNoContent {
		t.Fatalf("add mock to family expected 204, got %d: %s", status, string(resp))
	}

	status, resp = doJSON(t, client, http.MethodGet, srv.Server.URL+"/mocks/v1/families/"+family.FamilyID, nil, headers)
	if status != http.StatusOK {
		t.Fatalf("get family expected 200, got %d: %s", status, string(resp))
	}
	var fam struct {
		FamilyID      string `json:"familyId"`
		PrimaryMockID string `json:"primaryMockId"`
		Items         []struct {
			MockID    string `json:"mockId"`
			IsPrimary bool   `json:"isPrimary"`
		} `json:"items"`
	}
	if err := json.Unmarshal(resp, &fam); err != nil {
		t.Fatalf("decode family: %v", err)
	}
	if fam.PrimaryMockID != primaryID {
		t.Fatalf("primary mismatch: %+v", fam)
	}
	if len(fam.Items) != 2 {
		t.Fatalf("family items expected 2, got %d", len(fam.Items))
	}
	primaryFound := false
	secondaryFound := false
	for _, it := range fam.Items {
		if it.MockID == primaryID && it.IsPrimary {
			primaryFound = true
		}
		if it.MockID == otherID && !it.IsPrimary {
			secondaryFound = true
		}
	}
	if !primaryFound || !secondaryFound {
		t.Fatalf("family composition mismatch: %+v", fam.Items)
	}
}
