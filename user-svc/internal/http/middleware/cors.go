package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAgeSeconds    int
}

func WithCORS(cfg CORSConfig) func(http.Handler) http.Handler {
	if len(cfg.AllowedMethods) == 0 {
		cfg.AllowedMethods = []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPatch,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		}
	}
	if len(cfg.AllowedHeaders) == 0 {
		cfg.AllowedHeaders = []string{"Content-Type", "Authorization"}
	}
	if cfg.MaxAgeSeconds == 0 {
		cfg.MaxAgeSeconds = 3600
	}

	allowedMethods := strings.Join(cfg.AllowedMethods, ", ")
	allowedHeaders := strings.Join(cfg.AllowedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			if !isOriginAllowed(origin, cfg.AllowedOrigins) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			allowOrigin := origin
			if hasWildcard(cfg.AllowedOrigins) && !cfg.AllowCredentials {
				allowOrigin = "*"
			}

			headers := w.Header()
			headers.Set("Access-Control-Allow-Origin", allowOrigin)
			headers.Set("Vary", "Origin")
			headers.Set("Access-Control-Allow-Methods", allowedMethods)
			headers.Set("Access-Control-Allow-Headers", allowedHeaders)
			headers.Set("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAgeSeconds))
			if cfg.AllowCredentials {
				headers.Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isOriginAllowed(origin string, allowed []string) bool {
	if len(allowed) == 0 {
		return false
	}
	if hasWildcard(allowed) {
		return true
	}
	for _, value := range allowed {
		if strings.EqualFold(strings.TrimSpace(value), origin) {
			return true
		}
	}
	return false
}

func hasWildcard(allowed []string) bool {
	for _, value := range allowed {
		if strings.TrimSpace(value) == "*" {
			return true
		}
	}
	return false
}
