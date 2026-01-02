package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"mock-svc/internal/application"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, ErrorResponse{
		Code:    code,
		Message: msg,
	})
}

func writeServiceError(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, application.ErrInvalidRequest):
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
	case errors.Is(err, application.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
	case errors.Is(err, application.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "forbidden")
	case errors.Is(err, application.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "not found")
	case errors.Is(err, application.ErrConflict):
		writeError(w, http.StatusConflict, "CONFLICT", "conflict")
	case errors.Is(err, application.ErrValidation):
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION", "validation error")
	default:
		return false
	}
	return true
}

func writeInternalError(w http.ResponseWriter, err error) {
	slog.Error("request failed", "error", err)
	writeError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body")
		return false
	}
	return true
}
