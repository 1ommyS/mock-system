package handlers

import (
	"net/http"
	"strings"
	"time"
)

type ResourceRegisterRequest struct {
	ResourceType string `json:"resourceType"`
	ResourceID   string `json:"resourceId"`
	OwnerUserID  string `json:"ownerUserId"`
}

type ResourceRegisterResponse struct {
	ResourceType string `json:"resourceType"`
	ResourceID   string `json:"resourceId"`
	OwnerUserID  string `json:"ownerUserId"`
}

type CheckAccessRequest struct {
	ResourceType string `json:"resourceType"`
	ResourceID   string `json:"resourceId"`
	Action       string `json:"action"`
}

type CheckAccessResponse struct {
	Allowed             bool   `json:"allowed"`
	EffectivePermission string `json:"effectivePermission"`
	IsOwner             bool   `json:"isOwner"`
}

type GrantRequest struct {
	ResourceType  string `json:"resourceType"`
	ResourceID    string `json:"resourceId"`
	GranteeUserID string `json:"granteeUserId"`
	Permission    string `json:"permission"`
}

type GrantResponse struct {
	GrantID string `json:"grantId"`
}

type RevokeRequest struct {
	ResourceType  string `json:"resourceType"`
	ResourceID    string `json:"resourceId"`
	GranteeUserID string `json:"granteeUserId"`
}

type ListResourcesResponse struct {
	Items []ResourceItem `json:"items"`
}

type ResourceItem struct {
	ResourceID          string `json:"resourceId"`
	EffectivePermission string `json:"effectivePermission"`
	IsOwner             bool   `json:"isOwner"`
}

type ListGrantsResponse struct {
	Items []GrantItem `json:"items"`
}

type GrantItem struct {
	GranteeUserID string `json:"granteeUserId"`
	Permission    string `json:"permission"`
	CreatedAt     string `json:"createdAt"`
}

// RegisterResource godoc
// @Summary Register resource (ADMIN)
// @Tags authz
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ResourceRegisterRequest true "Register resource"
// @Success 201 {object} ResourceRegisterResponse
// @Success 200 {object} ResourceRegisterResponse
// @Failure 400 {object} ErrorResponse
// @Router /authz/v1/resources [post]
func (h *Handler) RegisterResource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req ResourceRegisterRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.ResourceType == "" || req.ResourceID == "" || req.OwnerUserID == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "resourceType, resourceId, ownerUserId are required")
		return
	}
	if !isValidResourceType(req.ResourceType) {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "unsupported resourceType")
		return
	}
	res, created, err := h.Services.RegisterResource(r.Context(), req.ResourceType, req.ResourceID, req.OwnerUserID)
	if err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, ResourceRegisterResponse{
		ResourceType: res.ResourceType,
		ResourceID:   res.ResourceID,
		OwnerUserID:  res.OwnerUserID,
	})
}

// CheckAccess godoc
// @Summary Check access for resource
// @Tags authz
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CheckAccessRequest true "Check access"
// @Success 200 {object} CheckAccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /authz/v1/check [post]
func (h *Handler) CheckAccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req CheckAccessRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.ResourceType == "" || req.ResourceID == "" || req.Action == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "resourceType, resourceId, action are required")
		return
	}
	if !isValidResourceType(req.ResourceType) {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "unsupported resourceType")
		return
	}
	userID, ok := userIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user")
		return
	}
	allowed, effective, isOwner, err := h.Services.CheckAccess(r.Context(), userID, req.ResourceType, req.ResourceID, req.Action)
	if err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, CheckAccessResponse{
		Allowed:             allowed,
		EffectivePermission: effective,
		IsOwner:             isOwner,
	})
}

// ListResources godoc
// @Summary List resources by permission
// @Tags authz
// @Security BearerAuth
// @Produce json
// @Param resourceType query string true "Resource type" Enums(mock,dsl_script)
// @Param minPermission query string false "Min permission" Enums(read,edit)
// @Success 200 {object} ListResourcesResponse
// @Router /authz/v1/resources [get]
func (h *Handler) ListResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	resourceType := r.URL.Query().Get("resourceType")
	minPermission := r.URL.Query().Get("minPermission")
	if resourceType == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "resourceType is required")
		return
	}
	if !isValidResourceType(resourceType) {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "unsupported resourceType")
		return
	}
	if minPermission != "" && !isValidMinPermission(minPermission) {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid minPermission")
		return
	}
	userID, ok := userIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user")
		return
	}
	items, err := h.Services.ListResources(r.Context(), userID, resourceType, minPermission)
	if err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	resp := ListResourcesResponse{Items: make([]ResourceItem, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, ResourceItem{
			ResourceID:          item.ResourceID,
			EffectivePermission: item.EffectivePermission,
			IsOwner:             item.IsOwner,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

// CreateGrant godoc
// @Summary Create grant (owner/admin)
// @Tags authz
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body GrantRequest true "Create grant"
// @Success 201 {object} GrantResponse
// @Failure 400 {object} ErrorResponse
// @Router /authz/v1/grants [post]
func (h *Handler) CreateGrant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req GrantRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.ResourceType == "" || req.ResourceID == "" || req.GranteeUserID == "" || req.Permission == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "resourceType, resourceId, granteeUserId, permission are required")
		return
	}
	if !isValidResourceType(req.ResourceType) {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "unsupported resourceType")
		return
	}
	if !isValidGrantPermission(req.Permission) {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid permission")
		return
	}
	userID, ok := userIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user")
		return
	}
	grant, err := h.Services.CreateGrant(r.Context(), userID, rolesFromContext(r), req.ResourceType, req.ResourceID, req.GranteeUserID, strings.ToUpper(req.Permission))
	if err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, GrantResponse{GrantID: grant.ID})
}

// DeleteGrant godoc
// @Summary Delete grant (owner/admin)
// @Tags authz
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body RevokeRequest true "Delete grant"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Router /authz/v1/grants [delete]
func (h *Handler) DeleteGrant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req RevokeRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.ResourceType == "" || req.ResourceID == "" || req.GranteeUserID == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "resourceType, resourceId, granteeUserId are required")
		return
	}
	if !isValidResourceType(req.ResourceType) {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "unsupported resourceType")
		return
	}
	userID, ok := userIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user")
		return
	}
	if err := h.Services.DeleteGrant(r.Context(), userID, rolesFromContext(r), req.ResourceType, req.ResourceID, req.GranteeUserID); err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListGrants godoc
// @Summary List grants for resource (owner/admin)
// @Tags authz
// @Security BearerAuth
// @Produce json
// @Param resourceType query string true "Resource type" Enums(mock,dsl_script)
// @Param resourceId query string true "Resource ID"
// @Success 200 {object} ListGrantsResponse
// @Router /authz/v1/grants [get]
func (h *Handler) ListGrants(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	resourceType := r.URL.Query().Get("resourceType")
	resourceID := r.URL.Query().Get("resourceId")
	if resourceType == "" || resourceID == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "resourceType and resourceId are required")
		return
	}
	if !isValidResourceType(resourceType) {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "unsupported resourceType")
		return
	}
	userID, ok := userIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user")
		return
	}
	grants, err := h.Services.ListGrants(r.Context(), userID, rolesFromContext(r), resourceType, resourceID)
	if err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	resp := ListGrantsResponse{Items: make([]GrantItem, 0, len(grants))}
	for _, grant := range grants {
		resp.Items = append(resp.Items, GrantItem{
			GranteeUserID: grant.UserID,
			Permission:    grant.Permission,
			CreatedAt:     grant.CreatedAt.Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func isValidResourceType(value string) bool {
	switch value {
	case "mock", "dsl_script":
		return true
	default:
		return false
	}
}

func isValidGrantPermission(value string) bool {
	switch strings.ToUpper(value) {
	case "READ", "EDIT":
		return true
	default:
		return false
	}
}

func isValidMinPermission(value string) bool {
	switch strings.ToLower(value) {
	case "read", "edit":
		return true
	default:
		return false
	}
}
