package middleware

import (
	"encoding/json"
	"net/http"

	"dsl-runner-svc/internal/application"
)

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}

func WithInternalSecret(secret, headerName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get(headerName) != secret {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(errorResponse{Code: application.ErrCodeUnauthorized, Message: "unauthorized"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
