package integration

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"dsl-runner-svc/internal/application"
	"dsl-runner-svc/internal/domain"
	"dsl-runner-svc/internal/dsl"
	"dsl-runner-svc/internal/infrastructure/postgres"
	"dsl-runner-svc/tests/testutil"
)

func TestClosestStoreFromPoints(t *testing.T) {
	pg := testutil.StartPostgres(t)
	defer pg.Close(t)

	testutil.ApplyMigrations(t, pg.DB, "../../migrations")

	store := postgres.NewStore(pg.DB)
	exec := dsl.NewExecutor()
	pub := &capturePublisher{}
	service := application.New(application.Repositories{Jobs: store.Jobs, Results: store.Results}, store, pub, exec, 3)
	handler := &application.JobHandler{Service: service}

	inputBytes, err := os.ReadFile("data/closest_store_input.json")
	if err != nil {
		t.Fatalf("read input: %v", err)
	}
	var msg domain.JobMessage
	if err := json.Unmarshal(inputBytes, &msg); err != nil {
		t.Fatalf("parse input: %v", err)
	}

	if err := handler.Process(context.Background(), msg); err != nil {
		t.Fatalf("process job: %v", err)
	}

	res, err := store.Results.GetByJobID(context.Background(), msg.JobID)
	if err != nil {
		t.Fatalf("get result: %v", err)
	}

	expectedBytes, err := os.ReadFile("data/closest_store_expected.json")
	if err != nil {
		t.Fatalf("read expected: %v", err)
	}

	actualCanon, err := dsl.CanonicalJSON(res.Result)
	if err != nil {
		t.Fatalf("canonical actual: %v", err)
	}
	expectedCanon, err := dsl.CanonicalJSON(expectedBytes)
	if err != nil {
		t.Fatalf("canonical expected: %v", err)
	}
	if actualCanon != expectedCanon {
		t.Fatalf("result mismatch\nactual: %s\nexpected: %s", actualCanon, expectedCanon)
	}

	if len(pub.messages) != 1 {
		t.Fatalf("expected 1 result message, got %d", len(pub.messages))
	}
}
