package handlers

import (
	"net/http"
	"strings"
)

type AdminUpdateUserRequest struct {
	Status string   `json:"status,omitempty"`
	Roles  []string `json:"roles,omitempty"`
}

type AdminUpdateUserResponse struct {
	UserID string   `json:"userId"`
	Status string   `json:"status"`
	Roles  []string `json:"roles"`
}

type AdminUpdateResourceRequest struct {
	OwnerUserID string `json:"ownerUserId,omitempty"`
}

type AdminUpdateResourceResponse struct {
	ResourceType string `json:"resourceType"`
	ResourceID   string `json:"resourceId"`
	OwnerUserID  string `json:"ownerUserId"`
}

// AdminUpdateUser godoc
// @Summary Admin update user (status/roles)
// @Tags admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param request body AdminUpdateUserRequest true "Update payload"
// @Success 200 {object} AdminUpdateUserResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/v1/admin/users/{userId} [patch]
func (h *Handler) AdminUpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID := strings.TrimPrefix(r.URL.Path, "/auth/v1/admin/users/")
	if userID == "" || userID == "/auth/v1/admin/users" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "userId is required in path")
		return
	}
	var req AdminUpdateUserRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	var status *string
	if req.Status != "" {
		status = &req.Status
	}
	user, roles, err := h.Services.UpdateUser(r.Context(), userID, status, req.Roles)
	if err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, AdminUpdateUserResponse{
		UserID: user.ID,
		Status: user.Status,
		Roles:  roles,
	})
}

// AdminUpdateResource godoc
// @Summary Admin change resource owner
// @Tags admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param resourceType path string true "Resource type" Enums(mock,dsl_script)
// @Param resourceId path string true "Resource ID"
// @Param request body AdminUpdateResourceRequest true "Update owner"
// @Success 200 {object} AdminUpdateResourceResponse
// @Failure 400 {object} ErrorResponse
// @Router /authz/v1/admin/resources/{resourceType}/{resourceId} [patch]
func (h *Handler) AdminUpdateResource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/authz/v1/admin/resources/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "resourceType and resourceId are required in path")
		return
	}
	if !isValidResourceType(parts[0]) {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "unsupported resourceType")
		return
	}
	var req AdminUpdateResourceRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.OwnerUserID == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "ownerUserId is required")
		return
	}
	res, err := h.Services.UpdateResourceOwner(r.Context(), parts[0], parts[1], req.OwnerUserID)
	if err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, AdminUpdateResourceResponse{
		ResourceType: res.ResourceType,
		ResourceID:   res.ResourceID,
		OwnerUserID:  res.OwnerUserID,
	})
}
