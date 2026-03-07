package dsl

import (
	"encoding/json"
	"testing"

	"dsl-runner-svc/internal/domain"
)

func TestParseScriptText(t *testing.T) {
	src := `
version 1

let basePath = input.base.request_match.path

repeat 2 {
  emit "gen-" + loop.index {
    requestMatch {
      method: input.base.request_match.method
      path: basePath + "/v" + loop.index
    }
    responseTemplate {
      status: 200
      body: {
        stableId: randUUID()
        n: randInt(1, 10)
      }
    }
    meta {
      from: "dsl"
      i: loop.index
    }
  }
}
`

	script, err := ParseScript([]byte(src))
	if err != nil {
		t.Fatalf("parse text script: %v", err)
	}
	if script.Version != 1 {
		t.Fatalf("version mismatch: %d", script.Version)
	}
	if len(script.Steps) != 2 {
		t.Fatalf("steps mismatch: %d", len(script.Steps))
	}
}

func TestExecutorRunTextDSL(t *testing.T) {
	src := `
version 1

repeat 2 {
  emit "gen-" + loop.index {
    requestMatch { method: "GET", path: "/v" + loop.index }
    responseTemplate { status: 200, body: { i: loop.index, id: randUUID() } }
    meta { from: "dsl", i: loop.index }
  }
}
`

	exec := NewExecutor()
	res, err := exec.Run(ExecInput{
		Seed:       "seed-text-1",
		ScriptText: src,
		BaseMock:   json.RawMessage(`{"request_match":{"method":"GET","path":"/base"}}`),
		Params:     json.RawMessage(`{}`),
		Limits: domain.JobLimits{
			TimeoutMs:         5000,
			MaxGeneratedMocks: 10,
			MaxResultBytes:    2_000_000,
		},
	})
	if err != nil {
		t.Fatalf("run text script: %v", err)
	}
	if len(res.Generated) != 2 {
		t.Fatalf("expected 2 generated mocks, got %d", len(res.Generated))
	}
	if res.Generated[0].Name != "gen-0" || res.Generated[1].Name != "gen-1" {
		t.Fatalf("unexpected names: %q, %q", res.Generated[0].Name, res.Generated[1].Name)
	}
}

func TestExecutorRunTextDSLForLoop(t *testing.T) {
	src := `
version 1

for item, idx in ["a", "b"] {
  emit item {
    requestMatch { method: "GET", path: "/v" + idx }
    responseTemplate { status: 200, body: { item: item, idx: idx } }
    meta { idx: idx }
  }
}
`

	exec := NewExecutor()
	res, err := exec.Run(ExecInput{
		Seed:       "seed-text-for",
		ScriptText: src,
		BaseMock:   json.RawMessage(`{}`),
		Params:     json.RawMessage(`{}`),
		Limits: domain.JobLimits{
			TimeoutMs:         5000,
			MaxGeneratedMocks: 10,
			MaxResultBytes:    2_000_000,
		},
	})
	if err != nil {
		t.Fatalf("run text for script: %v", err)
	}
	if len(res.Generated) != 2 {
		t.Fatalf("expected 2 generated mocks, got %d", len(res.Generated))
	}
	if res.Generated[0].Name != "a" || res.Generated[1].Name != "b" {
		t.Fatalf("unexpected names: %q, %q", res.Generated[0].Name, res.Generated[1].Name)
	}
}
