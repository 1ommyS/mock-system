package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"mock-svc/internal/application"
	"mock-svc/internal/domain"
)

type CreateGenerationRequest struct {
	BaseMockID  string          `json:"baseMockId"`
	DSLScriptID string          `json:"dslScriptId"`
	Params      json.RawMessage `json:"params,omitempty"`
	Mode        string          `json:"mode"`
}

type PreviewGenerationResponse struct {
	Planned []application.GeneratedMock `json:"planned"`
}

type ApplyGenerationResponse struct {
	GenerationID string `json:"generationId"`
	Status       string `json:"status"`
}

type GenerationResponse struct {
	GenerationID  string   `json:"generationId"`
	Status        string   `json:"status"`
	BaseMockID    string   `json:"baseMockId"`
	DSLScriptID   string   `json:"dslScriptId"`
	ResultMockIDs []string `json:"resultMockIds,omitempty"`
	Error         *string  `json:"error,omitempty"`
	CreatedAt     string   `json:"createdAt"`
	UpdatedAt     string   `json:"updatedAt"`
}

func (h *Handler) CreateGeneration(w http.ResponseWriter, r *http.Request) {
	var req CreateGenerationRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	userID, ok := userIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user")
		return
	}
	token, _ := tokenFromContext(r)
	input := application.GenerationInput{
		BaseMockID:  req.BaseMockID,
		DSLScriptID: req.DSLScriptID,
		Params:      req.Params,
	}
	mode := strings.ToLower(req.Mode)
	switch mode {
	case "preview":
		planned, err := h.Services.PreviewGeneration(r.Context(), token, userID, input)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, PreviewGenerationResponse{Planned: planned})
	case "apply":
		gen, err := h.Services.StartGeneration(r.Context(), token, userID, input)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}
		writeJSON(w, http.StatusAccepted, ApplyGenerationResponse{GenerationID: gen.ID, Status: gen.Status})
	default:
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid mode")
	}
}

func (h *Handler) GetGeneration(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/mocks/v1/generations/")
	if id == "" || id == "/mocks/v1/generations" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "generation id is required")
		return
	}
	userID, ok := userIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user")
		return
	}
	token, _ := tokenFromContext(r)
	gen, err := h.Services.GetGeneration(r.Context(), token, userID, rolesFromContext(r), id)
	if err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toGenerationResponse(gen))
}

func toGenerationResponse(gen domain.Generation) GenerationResponse {
	return GenerationResponse{
		GenerationID:  gen.ID,
		Status:        gen.Status,
		BaseMockID:    gen.BaseMockID,
		DSLScriptID:   gen.DSLScriptID,
		ResultMockIDs: gen.ResultMockIDs,
		Error:         gen.Error,
		CreatedAt:     gen.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     gen.UpdatedAt.Format(time.RFC3339),
	}
}
