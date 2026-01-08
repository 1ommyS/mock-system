package integration_test

import (
	"bytes"
	"encoding/json"
	"io"
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
)

type registerResponse struct {
	UserID string `json:"userId"`
}

type tokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

type meResponse struct {
	UserID string   `json:"userId"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	Status string   `json:"status"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func TestAuthFlow_RegisterLoginRefreshLogout(t *testing.T) {
	srv, cleanup := newAuthServer(t)
	defer cleanup()

	client := srv.Client()

	regBody := map[string]string{
		"email":    "user@example.com",
		"password": "secret123",
	}
	status, body := doJSON(t, client, http.MethodPost, srv.URL+"/auth/v1/register", regBody, nil)
	if status != http.StatusCreated {
		t.Fatalf("register expected 201, got %d: %s", status, string(body))
	}
	var regResp registerResponse
	if err := json.Unmarshal(body, &regResp); err != nil {
		t.Fatalf("decode register: %v", err)
	}

	status, body = doJSON(t, client, http.MethodPost, srv.URL+"/auth/v1/login", regBody, nil)
	if status != http.StatusOK {
		t.Fatalf("login expected 200, got %d: %s", status, string(body))
	}
	var loginResp tokenResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	if loginResp.AccessToken == "" || loginResp.RefreshToken == "" {
		t.Fatalf("tokens missing on login")
	}

	headers := map[string]string{
		"Authorization": "Bearer " + loginResp.AccessToken,
	}
	status, body = doJSON(t, client, http.MethodGet, srv.URL+"/auth/v1/me", nil, headers)
	if status != http.StatusOK {
		t.Fatalf("me expected 200, got %d: %s", status, string(body))
	}
	var me meResponse
	if err := json.Unmarshal(body, &me); err != nil {
		t.Fatalf("decode me: %v", err)
	}
	if me.UserID != regResp.UserID || me.Email != "user@example.com" || me.Status != "ACTIVE" {
		t.Fatalf("me data mismatch: %+v", me)
	}
	if !contains(me.Roles, "USER") {
		t.Fatalf("missing USER role: %+v", me.Roles)
	}

	refreshBody := map[string]string{"refreshToken": loginResp.RefreshToken}
	status, body = doJSON(t, client, http.MethodPost, srv.URL+"/auth/v1/token/refresh", refreshBody, nil)
	if status != http.StatusOK {
		t.Fatalf("refresh expected 200, got %d: %s", status, string(body))
	}
	var refreshResp tokenResponse
	if err := json.Unmarshal(body, &refreshResp); err != nil {
		t.Fatalf("decode refresh: %v", err)
	}
	if refreshResp.RefreshToken == "" || refreshResp.RefreshToken == loginResp.RefreshToken {
		t.Fatalf("refresh token not rotated")
	}

	headers = map[string]string{
		"Authorization": "Bearer " + refreshResp.AccessToken,
	}
	logoutBody := map[string]string{"refreshToken": refreshResp.RefreshToken}
	status, body = doJSON(t, client, http.MethodPost, srv.URL+"/auth/v1/logout", logoutBody, headers)
	if status != http.StatusNoContent {
		t.Fatalf("logout expected 204, got %d: %s", status, string(body))
	}

	status, body = doJSON(t, client, http.MethodPost, srv.URL+"/auth/v1/token/refresh", logoutBody, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("refresh after logout expected 401, got %d: %s", status, string(body))
	}
	var errResp errorResponse
	_ = json.Unmarshal(body, &errResp)
	if errResp.Code != "TOKEN_REVOKED" && errResp.Code != "TOKEN_EXPIRED" {
		t.Fatalf("unexpected error code: %s", errResp.Code)
	}
}

func TestAuth_RegisterDuplicate(t *testing.T) {
	srv, cleanup := newAuthServer(t)
	defer cleanup()

	client := srv.Client()
	regBody := map[string]string{
		"email":    "dup@example.com",
		"password": "secret123",
	}
	status, _ := doJSON(t, client, http.MethodPost, srv.URL+"/auth/v1/register", regBody, nil)
	if status != http.StatusCreated {
		t.Fatalf("register expected 201, got %d", status)
	}
	status, body := doJSON(t, client, http.MethodPost, srv.URL+"/auth/v1/register", regBody, nil)
	if status != http.StatusConflict {
		t.Fatalf("duplicate register expected 409, got %d: %s", status, string(body))
	}
}

func TestAuth_LoginInvalidPassword(t *testing.T) {
	srv, cleanup := newAuthServer(t)
	defer cleanup()

	client := srv.Client()
	regBody := map[string]string{
		"email":    "login@example.com",
		"password": "secret123",
	}
	status, _ := doJSON(t, client, http.MethodPost, srv.URL+"/auth/v1/register", regBody, nil)
	if status != http.StatusCreated {
		t.Fatalf("register expected 201, got %d", status)
	}

	badLogin := map[string]string{
		"email":    "login@example.com",
		"password": "wrong",
	}
	status, body := doJSON(t, client, http.MethodPost, srv.URL+"/auth/v1/login", badLogin, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("login invalid expected 401, got %d: %s", status, string(body))
	}
}

func TestAuth_RefreshReuseRevokes(t *testing.T) {
	srv, cleanup := newAuthServer(t)
	defer cleanup()

	client := srv.Client()
	regBody := map[string]string{
		"email":    "reuse@example.com",
		"password": "secret123",
	}
	status, _ := doJSON(t, client, http.MethodPost, srv.URL+"/auth/v1/register", regBody, nil)
	if status != http.StatusCreated {
		t.Fatalf("register expected 201, got %d", status)
	}

	status, body := doJSON(t, client, http.MethodPost, srv.URL+"/auth/v1/login", regBody, nil)
	if status != http.StatusOK {
		t.Fatalf("login expected 200, got %d: %s", status, string(body))
	}
	var loginResp tokenResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		t.Fatalf("decode login: %v", err)
	}

	refreshBody := map[string]string{"refreshToken": loginResp.RefreshToken}
	status, body = doJSON(t, client, http.MethodPost, srv.URL+"/auth/v1/token/refresh", refreshBody, nil)
	if status != http.StatusOK {
		t.Fatalf("refresh expected 200, got %d: %s", status, string(body))
	}

	status, body = doJSON(t, client, http.MethodPost, srv.URL+"/auth/v1/token/refresh", refreshBody, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("reuse refresh expected 401, got %d: %s", status, string(body))
	}
}

func newAuthServer(t *testing.T) (*httptest.Server, func()) {
	t.Helper()

	pg := testutil.StartPostgres(t)
	testutil.ApplyMigrations(t, pg.DB, "../../migrations")

	jwtSvc := auth.NewJWTService("test-secret", "user-svc", "user-svc")
	store := postgres.NewStore(pg.DB)
	repos := application.Repositories{
		Users:     store.Users,
		Roles:     store.Roles,
		Sessions:  store.Sessions,
		Resources: store.Resources,
		Grants:    store.Grants,
		Audit:     store.Audit,
	}
	services := application.New(repos, jwtSvc, 15*time.Minute, 30*24*time.Hour)
	handler := handlers.New(services)
	router := httpserver.NewRouter(
		handler,
		jwtSvc,
		middleware.CORSConfig{},
		"",
	)
	server := httptest.NewServer(router)

	cleanup := func() {
		server.Close()
		pg.Close(t)
	}

	return server, cleanup
}

func doJSON(t *testing.T, client *http.Client, method, url string, body any, headers map[string]string) (int, []byte) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, respBody
}

func contains(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}
