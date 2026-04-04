package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWaitForResultReturnsFailedJobError(t *testing.T) {
	const (
		secret = "test-secret"
		jobID  = "job-1"
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Internal-Secret") != secret {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/dslrunner/v1/jobs/" + jobID + "/result":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"code":"JOB_NOT_READY","message":"job not ready"}`))
		case "/dslrunner/v1/jobs/" + jobID:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"FAILED","error":{"code":"DSL_EVAL_ERROR","message":"let: dsl eval error"}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := NewDSLRunnerClient(DSLRunnerConfig{
		BaseURL:        srv.URL,
		InternalSecret: secret,
		PollInterval:   10 * time.Millisecond,
		PollTimeout:    200 * time.Millisecond,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := client.waitForResult(ctx, jobID)
	if err == nil {
		t.Fatalf("expected error for failed job")
	}
	if !strings.Contains(err.Error(), "DSL_EVAL_ERROR") {
		t.Fatalf("expected DSL_EVAL_ERROR in error, got: %v", err)
	}
}

