package httpserver

import (
	"net/http"
	"strings"

	"user-svc/internal/auth"
	"user-svc/internal/http/handlers"
	"user-svc/internal/http/middleware"
	"user-svc/internal/http/swagger"
)

func NewRouter(
	h *handlers.Handler,
	jwtSvc *auth.JWTService,
	corsConfig middleware.CORSConfig,
	internalSecret string,
) http.Handler {
	mux := http.NewServeMux()

	authMW := middleware.WithAuth(jwtSvc, internalSecret)
	adminMW := middleware.RequireRole("ADMIN")

	mux.HandleFunc("/healthz", h.Healthz)

	mux.Handle("/auth/v1/register", http.HandlerFunc(h.Register))
	mux.Handle("/auth/v1/login", http.HandlerFunc(h.Login))
	mux.Handle("/auth/v1/token/refresh", http.HandlerFunc(h.Refresh))
	mux.Handle("/auth/v1/logout", chain(http.HandlerFunc(h.Logout), authMW))
	mux.Handle("/auth/v1/me", chain(http.HandlerFunc(h.Me), authMW))

	mux.Handle("/auth/v1/admin/users/", chain(http.HandlerFunc(h.AdminUpdateUser), authMW, adminMW))

	mux.Handle("/authz/v1/resources", chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			if !hasRole(r, "ADMIN") {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			h.RegisterResource(w, r)
		case http.MethodGet:
			h.ListResources(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}), authMW))
	mux.Handle("/authz/v1/check", chain(http.HandlerFunc(h.CheckAccess), authMW))

	mux.Handle("/authz/v1/grants", chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.CreateGrant(w, r)
		case http.MethodDelete:
			h.DeleteGrant(w, r)
		case http.MethodGet:
			h.ListGrants(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}), authMW))

	mux.Handle("/authz/v1/admin/resources/", chain(http.HandlerFunc(h.AdminUpdateResource), authMW, adminMW))

	mux.Handle("/swagger/openapi.yaml", http.HandlerFunc(swagger.SpecHandler))
	mux.Handle("/swagger/", http.HandlerFunc(swagger.UIHandler))
	mux.Handle("/swagger", http.HandlerFunc(swagger.UIHandler))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = cleanPath(r.URL.Path)
		mux.ServeHTTP(w, r)
	})
	return middleware.WithCORS(corsConfig)(handler)
}

func chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func cleanPath(path string) string {
	if path == "" || path == "/" {
		return path
	}
	if strings.HasSuffix(path, "/") {
		return strings.TrimSuffix(path, "/")
	}
	return path
}

func hasRole(r *http.Request, role string) bool {
	for _, roleName := range middleware.RolesFromContext(r.Context()) {
		if roleName == role {
			return true
		}
	}
	return false
}
