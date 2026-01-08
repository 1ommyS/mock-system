package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"user-svc/internal/application"
	"user-svc/internal/auth"
	httpserver "user-svc/internal/http"
	"user-svc/internal/http/handlers"
	"user-svc/internal/http/middleware"
	"user-svc/internal/infrastructure/postgres"
	"user-svc/tests/testutil"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestHealthz(t *testing.T) {
	handler := handlers.New(nil)
	router := httpserver.NewRouter(
		handler,
		auth.NewJWTService("secret", "issuer", "aud"),
		middleware.CORSConfig{},
		"",
	)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHealthz_MethodNotAllowed(t *testing.T) {
	handler := handlers.New(nil)
	router := httpserver.NewRouter(
		handler,
		auth.NewJWTService("secret", "issuer", "aud"),
		middleware.CORSConfig{},
		"",
	)

	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestProtectedEndpoint_NoAuth(t *testing.T) {
	handler := handlers.New(nil)
	router := httpserver.NewRouter(
		handler,
		auth.NewJWTService("secret", "issuer", "aud"),
		middleware.CORSConfig{},
		"",
	)

	req := httptest.NewRequest(http.MethodGet, "/auth/v1/me", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", rr.Code)
	}
}

func TestProtectedEndpoint_WithBadToken(t *testing.T) {
	handler := handlers.New(nil)
	router := httpserver.NewRouter(
		handler,
		auth.NewJWTService("secret", "issuer", "aud"),
		middleware.CORSConfig{},
		"",
	)

	req := httptest.NewRequest(http.MethodGet, "/auth/v1/me", nil)
	req.Header.Set("Authorization", "Bearer invalid.token")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for bad token, got %d", rr.Code)
	}
}

func TestProtectedEndpoint_WithValidToken(t *testing.T) {
	db, mock, cleanup := testutil.NewMockDB(t)
	defer cleanup()

	jwtSvc := auth.NewJWTService("secret", "issuer", "aud")
	store := postgres.NewStore(db)
	repos := application.Repositories{
		Users:     store.Users,
		Roles:     store.Roles,
		Sessions:  store.Sessions,
		Resources: store.Resources,
		Grants:    store.Grants,
		Audit:     store.Audit,
	}
	svc := application.New(repos, jwtSvc, time.Minute, 30*24*time.Hour)
	handler := handlers.New(svc)
	token, _, err := jwtSvc.NewAccessToken("user-1", []string{"USER"}, time.Minute)
	if err != nil {
		t.Fatalf("token error: %v", err)
	}
	router := httpserver.NewRouter(
		handler,
		jwtSvc,
		middleware.CORSConfig{},
		"",
	)

	now := time.Now()
	userRows := sqlmock.NewRows([]string{
		"id", "email", "password_hash", "status", "created_at", "updated_at", "last_login_at",
	}).AddRow("user-1", "user@example.com", "hash", "ACTIVE", now, now, nil)
	mock.ExpectQuery(`SELECT id, email, password_hash, status, created_at, updated_at, last_login_at FROM users WHERE id = \$1`).
		WithArgs("user-1").
		WillReturnRows(userRows)
	mock.ExpectQuery(`SELECT role_name FROM user_roles WHERE user_id = \$1`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"role_name"}).AddRow("USER"))

	req := httptest.NewRequest(http.MethodGet, "/auth/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
