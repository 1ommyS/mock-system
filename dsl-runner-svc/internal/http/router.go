package httpserver

import (
	"net/http"
	"strings"

	"dsl-runner-svc/internal/http/handlers"
	"dsl-runner-svc/internal/http/middleware"
	"dsl-runner-svc/internal/http/swagger"
)

func NewRouter(h *handlers.Handler, internalSecret, headerName string) http.Handler {
	mux := http.NewServeMux()

	internalMW := middleware.WithInternalSecret(internalSecret, headerName)

	mux.HandleFunc("/healthz", h.Healthz)

	mux.Handle("/dslrunner/v1/jobs/", chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, "/result") {
				h.GetJobResult(w, r)
				return
			}
			h.GetJob(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}), internalMW))

	mux.Handle("/swagger/openapi.yaml", http.HandlerFunc(swagger.SpecHandler))
	mux.Handle("/swagger/", http.HandlerFunc(swagger.UIHandler))
	mux.Handle("/swagger", http.HandlerFunc(swagger.UIHandler))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = cleanPath(r.URL.Path)
		mux.ServeHTTP(w, r)
	})
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
