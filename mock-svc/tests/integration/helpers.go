package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"mock-svc/internal/application"
	httpserver "mock-svc/internal/http"
	"mock-svc/internal/http/handlers"
	"mock-svc/internal/http/middleware"
	"mock-svc/internal/infrastructure/postgres"
	"mock-svc/tests/testutil"

	"github.com/jmoiron/sqlx"
)

type fakeAuthClient struct {
	mu          sync.Mutex
	resources   map[string]map[string]struct{}
	alwaysAllow bool
}

func newFakeAuthClient(alwaysAllow bool) *fakeAuthClient {
	return &fakeAuthClient{
		resources:   make(map[string]map[string]struct{}),
		alwaysAllow: alwaysAllow,
	}
}

func (c *fakeAuthClient) CheckAccess(_ context.Context, _ string, resourceType, resourceID, _ string) (bool, string, bool, error) {
	if c.alwaysAllow {
		return true, "EDIT", true, nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if ids, ok := c.resources[resourceType]; ok {
		if _, exists := ids[resourceID]; exists {
			return true, "EDIT", true, nil
		}
	}
	return false, "NONE", false, nil
}

func (c *fakeAuthClient) ListResources(_ context.Context, _ string, resourceType, _ string) ([]application.ResourceAccess, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ids := c.resources[resourceType]
	items := make([]application.ResourceAccess, 0, len(ids))
	for id := range ids {
		items = append(items, application.ResourceAccess{
			ResourceID:          id,
			EffectivePermission: "EDIT",
			IsOwner:             true,
		})
	}
	return items, nil
}

func (c *fakeAuthClient) RegisterResource(_ context.Context, resourceType, resourceID, _ string) error {
	c.AddResource(resourceType, resourceID)
	return nil
}

func (c *fakeAuthClient) AddResource(resourceType, resourceID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.resources[resourceType]; !ok {
		c.resources[resourceType] = make(map[string]struct{})
	}
	c.resources[resourceType][resourceID] = struct{}{}
}

type fakeDSLRunner struct {
	mu      sync.Mutex
	planned []application.GeneratedMock
	derived []application.GeneratedMock
}

func (f *fakeDSLRunner) Preview(_ context.Context, _ application.DSLBaseMock, _ string, _ string, _ json.RawMessage) ([]application.GeneratedMock, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]application.GeneratedMock{}, f.planned...), nil
}

func (f *fakeDSLRunner) Apply(_ context.Context, _ application.DSLBaseMock, _ string, _ string, _ json.RawMessage) ([]application.GeneratedMock, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]application.GeneratedMock{}, f.derived...), nil
}

func (f *fakeDSLRunner) SetPreview(mocks []application.GeneratedMock) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.planned = mocks
}

func (f *fakeDSLRunner) SetDerived(mocks []application.GeneratedMock) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.derived = mocks
}

type testServer struct {
	Server    *httptest.Server
	Auth      *fakeAuthClient
	DSL       *fakeDSLRunner
	DB        *sqlx.DB
	CleanupFn func()
}

func newMockServer(t *testing.T) *testServer {
	t.Helper()

	pg := testutil.StartPostgres(t)
	testutil.ApplyMigrations(t, pg.DB, "../../migrations")

	authClient := newFakeAuthClient(false)
	dsl := &fakeDSLRunner{}

	store := postgres.NewStore(pg.DB)
	repos := application.Repositories{
		Mocks:       store.Mocks,
		Families:    store.Families,
		Derived:     store.Derived,
		Generations: store.Generations,
		MockSearch:  store.MockSearch,
		Outbox:      store.Outbox,
	}
	services := application.New(repos, store, authClient, dsl, 3)
	handler := handlers.New(services)
	router := httpserver.NewRouter(handler, middleware.CORSConfig{})
	server := httptest.NewServer(router)

	return &testServer{
		Server: server,
		Auth:   authClient,
		DSL:    dsl,
		DB:     pg.DB,
		CleanupFn: func() {
			server.Close()
			pg.Close(t)
		},
	}
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
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
