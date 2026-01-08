package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"user-svc/internal/application"
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
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Некорректный запрос.")
	case errors.Is(err, application.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Неверные учетные данные.")
	case errors.Is(err, application.ErrTokenExpired):
		writeError(w, http.StatusUnauthorized, "TOKEN_EXPIRED", "Срок действия токена истек.")
	case errors.Is(err, application.ErrTokenRevoked):
		writeError(w, http.StatusUnauthorized, "TOKEN_REVOKED", "Токен отозван.")
	case errors.Is(err, application.ErrInvalidToken):
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Некорректный токен.")
	case errors.Is(err, application.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Доступ запрещен.")
	case errors.Is(err, application.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Не найдено.")
	case errors.Is(err, application.ErrConflict):
		writeError(w, http.StatusConflict, "CONFLICT", "Запись уже существует.")
	default:
		return false
	}
	return true
}

func writeInternalError(w http.ResponseWriter, err error) {
	slog.Error("request failed", "error", err)
	writeError(w, http.StatusInternalServerError, "INTERNAL", "Внутренняя ошибка сервера.")
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Некорректный JSON в теле запроса.")
		return false
	}
	return true
}
