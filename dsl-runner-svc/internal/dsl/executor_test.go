package dsl

import (
	"encoding/json"
	"testing"

	"dsl-runner-svc/internal/domain"
)

func TestDeterminism(t *testing.T) {
	script := `{"version":1,"steps":[{"op":"repeat","count":3,"do":[{"op":"emit","name":{"$fn":"concat","args":["gen-",{"$":"loop.index"}]},"requestMatch":{"path":{"$fn":"concat","args":["/v",{"$":"loop.index"}]}},"responseTemplate":{"body":{"id":{"$fn":"randUUID","args":[]},"n":{"$fn":"randInt","args":[1,10]}}},"meta":{"from":"dsl"}}]}]}`
	input := ExecInput{
		Seed:       "seed-1",
		ScriptText: script,
		BaseMock:   json.RawMessage(`{"request_match":{"path":"/base"}}`),
		Params:     json.RawMessage(`{}`),
		Limits: domain.JobLimits{
			TimeoutMs:         5000,
			MaxGeneratedMocks: 10,
			MaxResultBytes:    2000000,
		},
	}

	exec := NewExecutor()
	res1, err := exec.Run(input)
	if err != nil {
		t.Fatalf("run1 error: %v", err)
	}
	res2, err := exec.Run(input)
	if err != nil {
		t.Fatalf("run2 error: %v", err)
	}
	b1, _ := json.Marshal(res1.Generated)
	b2, _ := json.Marshal(res2.Generated)
	if string(b1) != string(b2) {
		t.Fatalf("results differ: %s vs %s", string(b1), string(b2))
	}
}
