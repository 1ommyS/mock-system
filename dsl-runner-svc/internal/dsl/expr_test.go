package dsl

import (
	"encoding/json"
	"testing"
)

func TestResolvePath(t *testing.T) {
	ctx := &EvalContext{
		InputBase: map[string]any{
			"request_match": map[string]any{
				"path": "/foo",
			},
		},
		InputParams: map[string]any{
			"list": []any{"a", "b"},
		},
		Vars: map[string]any{
			"x": 10,
		},
	}

	val, err := ResolvePath(ctx, "input.base.request_match.path")
	if err != nil || val != "/foo" {
		t.Fatalf("unexpected path value: %v, err=%v", val, err)
	}

	val, err = ResolvePath(ctx, "input.params.list[1]")
	if err != nil || val != "b" {
		t.Fatalf("unexpected path value: %v, err=%v", val, err)
	}

	val, err = ResolvePath(ctx, "vars.x")
	if err != nil || val != 10 {
		t.Fatalf("unexpected vars value: %v, err=%v", val, err)
	}
}

func TestEvalExprConcat(t *testing.T) {
	ctx := &EvalContext{Vars: map[string]any{"n": 3}}
	funcs := DefaultFuncs()
	expr := json.RawMessage(`{"$fn":"concat","args":["a", {"$":"vars.n"}]}`)
	val, err := EvalExpr(ctx, funcs, expr)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	if val != "a3" {
		t.Fatalf("unexpected concat: %v", val)
	}
}
