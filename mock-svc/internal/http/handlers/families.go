package handlers

import (
	"net/http"
	"strings"
)

type CreateFamilyRequest struct {
	PrimaryMockID string `json:"primaryMockId"`
}

type CreateFamilyResponse struct {
	FamilyID string `json:"familyId"`
}

type AddFamilyMockRequest struct {
	MockID string `json:"mockId"`
}

type FamilyResponse struct {
	FamilyID      string               `json:"familyId"`
	PrimaryMockID string               `json:"primaryMockId"`
	Items         []FamilyItemResponse `json:"items"`
}

type FamilyItemResponse struct {
	MockID            string  `json:"mockId"`
	IsPrimary         bool    `json:"isPrimary"`
	DerivedFromBaseID *string `json:"derivedFromBaseId,omitempty"`
	ViaDSLScriptID    *string `json:"viaDslScriptId,omitempty"`
}

type UpdateFamilyPrimaryRequest struct {
	NewPrimaryMockID string `json:"newPrimaryMockId"`
}

type UpdateFamilyPrimaryResponse struct {
	FamilyID      string `json:"familyId"`
	PrimaryMockID string `json:"primaryMockId"`
}

func (h *Handler) CreateFamily(w http.ResponseWriter, r *http.Request) {
	var req CreateFamilyRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	token, _ := tokenFromContext(r)
	family, err := h.Services.CreateFamily(r.Context(), token, req.PrimaryMockID)
	if err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, CreateFamilyResponse{FamilyID: family.ID})
}

func (h *Handler) AddFamilyMock(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/mocks/v1/families/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] != "mocks" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "familyId is required in path")
		return
	}
	var req AddFamilyMockRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	userID, ok := userIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user")
		return
	}
	token, _ := tokenFromContext(r)
	if err := h.Services.AddFamilyMock(r.Context(), token, userID, rolesFromContext(r), parts[0], req.MockID); err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetFamily(w http.ResponseWriter, r *http.Request) {
	familyID := strings.TrimPrefix(r.URL.Path, "/mocks/v1/families/")
	if familyID == "" || strings.Contains(familyID, "/") {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "familyId is required in path")
		return
	}
	token, _ := tokenFromContext(r)
	family, err := h.Services.GetFamily(r.Context(), token, familyID)
	if err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	items := make([]FamilyItemResponse, 0, len(family.Items))
	for _, item := range family.Items {
		items = append(items, FamilyItemResponse{
			MockID:            item.MockID,
			IsPrimary:         item.IsPrimary,
			DerivedFromBaseID: item.DerivedFromBaseID,
			ViaDSLScriptID:    item.ViaDSLScriptID,
		})
	}
	writeJSON(w, http.StatusOK, FamilyResponse{
		FamilyID:      family.FamilyID,
		PrimaryMockID: family.PrimaryMockID,
		Items:         items,
	})
}

func (h *Handler) UpdateFamilyPrimary(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/mocks/v1/families/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] != "primary" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "familyId is required in path")
		return
	}
	var req UpdateFamilyPrimaryRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	userID, ok := userIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user")
		return
	}
	token, _ := tokenFromContext(r)
	family, err := h.Services.UpdateFamilyPrimary(r.Context(), token, userID, rolesFromContext(r), parts[0], req.NewPrimaryMockID)
	if err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, UpdateFamilyPrimaryResponse{FamilyID: family.ID, PrimaryMockID: family.PrimaryMockID})
}
