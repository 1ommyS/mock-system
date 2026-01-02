package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"dsl-runner-svc/internal/application"
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
	writeJSON(w, status, ErrorResponse{Code: code, Message: msg})
}

func writeServiceError(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, application.ErrInvalidRequest):
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
	case errors.Is(err, application.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, application.ErrCodeUnauthorized, "unauthorized")
	case errors.Is(err, application.ErrNotFound):
		writeError(w, http.StatusNotFound, application.ErrCodeJobNotFound, "job not found")
	case errors.Is(err, application.ErrJobNotReady):
		writeError(w, http.StatusConflict, application.ErrCodeJobNotReady, "job not ready")
	default:
		return false
	}
	return true
}

func writeInternalError(w http.ResponseWriter, err error) {
	slog.Error("request failed", "error", err)
	writeError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
}
