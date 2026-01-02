package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mock-svc/internal/application"
	"mock-svc/internal/domain"
)

type CreateMockRequest struct {
	Name             string          `json:"name"`
	Description      *string         `json:"description,omitempty"`
	RequestMatch     json.RawMessage `json:"requestMatch"`
	ResponseTemplate json.RawMessage `json:"responseTemplate"`
	Enabled          *bool           `json:"enabled,omitempty"`
}

type CreateMockResponse struct {
	ID string `json:"id"`
}

type UpdateMockRequest struct {
	Name             *string          `json:"name,omitempty"`
	Description      OptionalString   `json:"description,omitempty"`
	RequestMatch     *json.RawMessage `json:"requestMatch,omitempty"`
	ResponseTemplate *json.RawMessage `json:"responseTemplate,omitempty"`
	Enabled          *bool            `json:"enabled,omitempty"`
}

type MockResponse struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Description      *string         `json:"description,omitempty"`
	RequestMatch     json.RawMessage `json:"requestMatch"`
	ResponseTemplate json.RawMessage `json:"responseTemplate"`
	Enabled          bool            `json:"enabled"`
	FamilyID         *string         `json:"familyId,omitempty"`
	CreatedBy        string          `json:"createdBy"`
	GeneratedBy      *string         `json:"generatedBy,omitempty"`
	CreatedAt        string          `json:"createdAt"`
	UpdatedAt        string          `json:"updatedAt"`
	DeletedAt        *string         `json:"deletedAt,omitempty"`
}

type ListMocksResponse struct {
	Items []MockResponse `json:"items"`
}

func (h *Handler) CreateMock(w http.ResponseWriter, r *http.Request) {
	var req CreateMockRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	userID, ok := userIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user")
		return
	}
	token, _ := tokenFromContext(r)
	mock, err := h.Services.CreateMock(r.Context(), userID, token, application.CreateMockInput{
		Name:             req.Name,
		Description:      req.Description,
		RequestMatch:     req.RequestMatch,
		ResponseTemplate: req.ResponseTemplate,
		Enabled:          req.Enabled,
	})
	if err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, CreateMockResponse{ID: mock.ID})
}

func (h *Handler) GetMock(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/mocks/v1/mocks/")
	if id == "" || id == "/mocks/v1/mocks" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "mock id is required")
		return
	}
	token, _ := tokenFromContext(r)
	mock, err := h.Services.GetMock(r.Context(), token, id)
	if err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toMockResponse(mock))
}

func (h *Handler) UpdateMock(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/mocks/v1/mocks/")
	if id == "" || id == "/mocks/v1/mocks" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "mock id is required")
		return
	}
	var req UpdateMockRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	patch := domain.MockPatch{
		Name:             req.Name,
		RequestMatch:     req.RequestMatch,
		ResponseTemplate: req.ResponseTemplate,
		Enabled:          req.Enabled,
	}
	if req.Description.Set {
		patch.Description = &req.Description.Value
	}
	token, _ := tokenFromContext(r)
	mock, err := h.Services.UpdateMock(r.Context(), token, id, patch)
	if err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toMockResponse(mock))
}

func (h *Handler) DeleteMock(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/mocks/v1/mocks/")
	if id == "" || id == "/mocks/v1/mocks" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "mock id is required")
		return
	}
	token, _ := tokenFromContext(r)
	if err := h.Services.DeleteMock(r.Context(), token, id); err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListMocks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filters := domain.MockSearchFilters{
		Query: query.Get("q"),
	}
	if familyID := query.Get("familyId"); familyID != "" {
		filters.FamilyID = &familyID
	}
	if enabledStr := query.Get("enabled"); enabledStr != "" {
		value, err := strconv.ParseBool(enabledStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid enabled")
			return
		}
		filters.Enabled = &value
	}
	if method := query.Get("method"); method != "" {
		filters.Method = &method
	}
	if path := query.Get("path"); path != "" {
		filters.Path = &path
	}
	if sort := query.Get("sort"); sort != "" && sort != "updatedAtDesc" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid sort")
		return
	}
	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid limit")
			return
		}
		filters.Limit = limit
	}
	if offsetStr := query.Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid offset")
			return
		}
		filters.Offset = offset
	}

	token, _ := tokenFromContext(r)
	items, err := h.Services.ListMocks(r.Context(), token, filters)
	if err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	resp := ListMocksResponse{Items: make([]MockResponse, 0, len(items))}
	for _, mock := range items {
		resp.Items = append(resp.Items, toMockResponse(mock))
	}
	writeJSON(w, http.StatusOK, resp)
}

func toMockResponse(mock domain.Mock) MockResponse {
	var deletedAt *string
	if mock.DeletedAt != nil {
		value := mock.DeletedAt.Format(time.RFC3339)
		deletedAt = &value
	}
	return MockResponse{
		ID:               mock.ID,
		Name:             mock.Name,
		Description:      mock.Description,
		RequestMatch:     mock.RequestMatch,
		ResponseTemplate: mock.ResponseTemplate,
		Enabled:          mock.Enabled,
		FamilyID:         mock.FamilyID,
		CreatedBy:        mock.CreatedBy,
		GeneratedBy:      mock.GeneratedBy,
		CreatedAt:        mock.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        mock.UpdatedAt.Format(time.RFC3339),
		DeletedAt:        deletedAt,
	}
}
