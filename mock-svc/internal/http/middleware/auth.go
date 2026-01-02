package middleware

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey string

const (
	userIDKey ctxKey = "userID"
	rolesKey  ctxKey = "roles"
	tokenKey  ctxKey = "token"
)

func WithContext() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			token := ""
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
			userID := r.Header.Get("X-User-Id")
			roles := parseRolesHeader(r.Header.Get("X-Roles"))
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			ctx = context.WithValue(ctx, rolesKey, roles)
			ctx = context.WithValue(ctx, tokenKey, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles := RolesFromContext(r.Context())
			for _, roleName := range roles {
				if roleName == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			w.WriteHeader(http.StatusForbidden)
		})
	}
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(userIDKey)
	if v == nil {
		return "", false
	}
	id, ok := v.(string)
	return id, ok
}

func RolesFromContext(ctx context.Context) []string {
	v := ctx.Value(rolesKey)
	if v == nil {
		return nil
	}
	roles, _ := v.([]string)
	return roles
}

func TokenFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(tokenKey)
	if v == nil {
		return "", false
	}
	token, ok := v.(string)
	return token, ok
}

func parseRolesHeader(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	roles := make([]string, 0, len(parts))
	for _, part := range parts {
		role := strings.TrimSpace(part)
		if role == "" {
			continue
		}
		roles = append(roles, role)
	}
	return roles
}
