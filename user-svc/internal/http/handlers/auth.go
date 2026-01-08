package handlers

import (
	"errors"
	"net"
	"net/http"
	"strings"

	"user-svc/internal/application"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name,omitempty"`
}

type RegisterResponse struct {
	UserID string `json:"userId"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type MeResponse struct {
	UserID string   `json:"userId"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	Status string   `json:"status"`
}

// Register godoc
// @Summary Register user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register payload"
// @Success 201 {object} RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/v1/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req RegisterRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Email и пароль обязательны.")
		return
	}
	userID, err := h.Services.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, application.ErrConflict) {
			writeError(w, http.StatusConflict, "CONFLICT", "Аккаунт с таким email уже существует.")
			return
		}
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, RegisterResponse{UserID: userID})
}

// Login godoc
// @Summary Login user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login payload"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/v1/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req LoginRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Email и пароль обязательны.")
		return
	}
	ip := clientIP(r)
	ua := r.UserAgent()
	tokenPair, err := h.Services.Login(r.Context(), req.Email, req.Password, strPtr(ip), strPtr(ua))
	if err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	})
}

// Refresh godoc
// @Summary Refresh access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh payload"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/v1/token/refresh [post]
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req RefreshRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Требуется refreshToken.")
		return
	}
	tokenPair, err := h.Services.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	})
}

// Logout godoc
// @Summary Logout user (revoke refresh)
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "Logout payload"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Router /auth/v1/logout [post]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req LogoutRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Требуется refreshToken.")
		return
	}
	if err := h.Services.Logout(r.Context(), req.RefreshToken); err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Me godoc
// @Summary Current user
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} MeResponse
// @Router /auth/v1/me [get]
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, ok := userIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация.")
		return
	}
	user, roles, err := h.Services.Me(r.Context(), userID)
	if err != nil {
		if writeServiceError(w, err) {
			return
		}
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, MeResponse{
		UserID: user.ID,
		Email:  user.Email,
		Roles:  roles,
		Status: user.Status,
	})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func strPtr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
