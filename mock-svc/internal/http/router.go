package httpserver

import (
	"net/http"
	"strings"

	"mock-svc/internal/http/handlers"
	"mock-svc/internal/http/middleware"
	"mock-svc/internal/http/swagger"
)

func NewRouter(h *handlers.Handler, corsConfig middleware.CORSConfig) http.Handler {
	mux := http.NewServeMux()

	authMW := middleware.WithContext()

	mux.HandleFunc("/healthz", h.Healthz)

	mux.Handle("/mocks/v1/mocks", chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.CreateMock(w, r)
		case http.MethodGet:
			h.ListMocks(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}), authMW))
	mux.Handle("/mocks/v1/mocks/", chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetMock(w, r)
		case http.MethodPatch:
			h.UpdateMock(w, r)
		case http.MethodDelete:
			h.DeleteMock(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}), authMW))

	mux.Handle("/mocks/v1/families", chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.CreateFamily(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}), authMW))
	mux.Handle("/mocks/v1/families/", chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.AddFamilyMock(w, r)
		case http.MethodGet:
			h.GetFamily(w, r)
		case http.MethodPatch:
			h.UpdateFamilyPrimary(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}), authMW))

	mux.Handle("/mocks/v1/generations", chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.CreateGeneration(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}), authMW))
	mux.Handle("/mocks/v1/generations/", chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetGeneration(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}), authMW))

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
