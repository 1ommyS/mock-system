package handlers

import (
	"net/http"

	"mock-svc/internal/http/middleware"
)

func userIDFromContext(r *http.Request) (string, bool) {
	return middleware.UserIDFromContext(r.Context())
}

func rolesFromContext(r *http.Request) []string {
	return middleware.RolesFromContext(r.Context())
}

func tokenFromContext(r *http.Request) (string, bool) {
	return middleware.TokenFromContext(r.Context())
}
