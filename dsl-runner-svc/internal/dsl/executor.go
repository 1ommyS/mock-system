package dsl

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"dsl-runner-svc/internal/domain"
)

type ExecInput struct {
	Seed       string
	ScriptText string
	BaseMock   json.RawMessage
	Params     json.RawMessage
	Limits     domain.JobLimits
}

type ExecResult struct {
	ScriptHash string
	InputHash  string
	Generated  []domain.GeneratedMockSpec
	Warnings   []string
}

type Executor struct {
	Funcs FuncRegistry
}

func NewExecutor() *Executor {
	return &Executor{Funcs: DefaultFuncs()}
}

func (e *Executor) Run(input ExecInput) (ExecResult, error) {
	if input.ScriptText == "" {
		return ExecResult{}, fmt.Errorf("empty script: %w", ErrValidation)
	}
	if input.Seed == "" {
		return ExecResult{}, fmt.Errorf("empty seed: %w", ErrValidation)
	}

	scriptHash := Sha256Hex([]byte(input.ScriptText))
	baseCanon, err := CanonicalJSON(input.BaseMock)
	if err != nil {
		return ExecResult{}, fmt.Errorf("canonical base: %w", ErrValidation)
	}
	paramsCanon, err := CanonicalJSON(input.Params)
	if err != nil {
		return ExecResult{}, fmt.Errorf("canonical params: %w", ErrValidation)
	}
	inputHash := Sha256Hex([]byte(baseCanon + paramsCanon + input.Seed + scriptHash))

	script, err := ParseScript([]byte(input.ScriptText))
	if err != nil {
		return ExecResult{}, fmt.Errorf("parse: %w", ErrParse)
	}

	ctx := &EvalContext{
		Vars: make(map[string]any),
	}
	if len(input.BaseMock) > 0 {
		if err := json.Unmarshal(input.BaseMock, &ctx.InputBase); err != nil {
			return ExecResult{}, fmt.Errorf("baseMock decode: %w", ErrValidation)
		}
	}
	if len(input.Params) > 0 {
		if err := json.Unmarshal(input.Params, &ctx.InputParams); err != nil {
			return ExecResult{}, fmt.Errorf("params decode: %w", ErrValidation)
		}
	}

	maxMocks := enforceMax(input.Limits.MaxGeneratedMocks, 1000, 200)
	maxBytes := enforceMax(input.Limits.MaxResultBytes, 10_000_000, 2_000_000)
	timeoutMs := enforceMax(input.Limits.TimeoutMs, 15000, 5000)

	exec := &executorState{
		ctx:         ctx,
		funcs:       e.Funcs,
		seed:        input.Seed,
		scriptHash:  scriptHash,
		maxMocks:    maxMocks,
		maxBytes:    maxBytes,
		resultBytes: 0,
	}

	if err := withTimeout(timeoutMs, func() error { return exec.runSteps(script.Steps) }); err != nil {
		return ExecResult{}, err
	}

	return ExecResult{
		ScriptHash: scriptHash,
		InputHash:  inputHash,
		Generated:  exec.generated,
		Warnings:   nil,
	}, nil
}

type executorState struct {
	ctx         *EvalContext
	funcs       FuncRegistry
	seed        string
	scriptHash  string
	emitted     int
	maxMocks    int
	maxBytes    int
	resultBytes int
	generated   []domain.GeneratedMockSpec
}

func (e *executorState) runSteps(steps []Step) error {
	for _, step := range steps {
		switch s := step.(type) {
		case LetStep:
			val, err := EvalExpr(e.ctx, e.funcs, s.Value)
			if err != nil {
				return fmt.Errorf("let: %w", ErrEval)
			}
			e.ctx.Vars[s.Name] = val
		case IfStep:
			condVal, err := EvalExpr(e.ctx, e.funcs, s.Cond)
			if err != nil {
				return fmt.Errorf("if: %w", ErrEval)
			}
			if truthy(condVal) {
				if err := e.runSteps(s.Then); err != nil {
					return err
				}
			} else if len(s.Else) > 0 {
				if err := e.runSteps(s.Else); err != nil {
					return err
				}
			}
		case RepeatStep:
			countVal, err := EvalExpr(e.ctx, e.funcs, s.Count)
			if err != nil {
				return fmt.Errorf("repeat: %w", ErrEval)
			}
			count, err := toIntStrict(countVal)
			if err != nil || count < 0 {
				return fmt.Errorf("repeat count invalid: %w", ErrValidation)
			}
			for i := 0; i < count; i++ {
				iCopy := i
				e.ctx.LoopIndex = &iCopy
				if err := e.runSteps(s.Do); err != nil {
					return err
				}
			}
			e.ctx.LoopIndex = nil
		case ForEachStep:
			itemsVal, err := EvalExpr(e.ctx, e.funcs, s.Items)
			if err != nil {
				return fmt.Errorf("forEach: %w", ErrEval)
			}
			items, ok := itemsVal.([]any)
			if !ok {
				return fmt.Errorf("forEach items must be array: %w", ErrValidation)
			}
			for i, item := range items {
				iCopy := i
				e.ctx.LoopIndex = &iCopy
				e.ctx.LoopItem = item
				if err := e.runSteps(s.Do); err != nil {
					return err
				}
			}
			e.ctx.LoopIndex = nil
			e.ctx.LoopItem = nil
		case EmitStep:
			if err := e.emit(s); err != nil {
				return err
			}
		case AssertStep:
			condVal, err := EvalExpr(e.ctx, e.funcs, s.Cond)
			if err != nil {
				return fmt.Errorf("assert: %w", ErrEval)
			}
			if !truthy(condVal) {
				return fmt.Errorf("assert failed: %s: %w", s.Error, ErrValidation)
			}
		default:
			return fmt.Errorf("unknown step: %w", ErrValidation)
		}
	}
	return nil
}

func (e *executorState) emit(step EmitStep) error {
	if e.emitted >= e.maxMocks {
		return fmt.Errorf("too many mocks: %w", ErrTooManyMocks)
	}
	derivedSeed := Sha256Hex([]byte(fmt.Sprintf("%s:%s:%d", e.seed, e.scriptHash, e.emitted)))
	e.ctx.Rand = NewRand(derivedSeed)
	e.ctx.AllowRandom = true

	nameVal, err := EvalExpr(e.ctx, e.funcs, step.Name)
	if err != nil {
		return fmt.Errorf("emit name: %w", ErrEval)
	}
	name, ok := nameVal.(string)
	if !ok || name == "" {
		return fmt.Errorf("emit name must be string: %w", ErrValidation)
	}

	requestVal, err := EvalExpr(e.ctx, e.funcs, step.RequestMatch)
	if err != nil {
		return fmt.Errorf("emit request: %w", ErrEval)
	}
	responseVal, err := EvalExpr(e.ctx, e.funcs, step.ResponseTemplate)
	if err != nil {
		return fmt.Errorf("emit response: %w", ErrEval)
	}
	metaVal, err := EvalExpr(e.ctx, e.funcs, step.Meta)
	if err != nil {
		return fmt.Errorf("emit meta: %w", ErrEval)
	}

	requestBytes, err := json.Marshal(requestVal)
	if err != nil {
		return fmt.Errorf("marshal request: %w", ErrRuntime)
	}
	responseBytes, err := json.Marshal(responseVal)
	if err != nil {
		return fmt.Errorf("marshal response: %w", ErrRuntime)
	}
	metaBytes, err := json.Marshal(metaVal)
	if err != nil {
		return fmt.Errorf("marshal meta: %w", ErrRuntime)
	}

	item := domain.GeneratedMockSpec{
		Name:             name,
		RequestMatch:     requestBytes,
		ResponseTemplate: responseBytes,
		Meta:             metaBytes,
		DerivedSeed:      derivedSeed,
	}

	chunk, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshal result: %w", ErrRuntime)
	}
	e.resultBytes += len(chunk)
	if e.resultBytes > e.maxBytes {
		return fmt.Errorf("result too large: %w", ErrResultTooLarge)
	}

	e.generated = append(e.generated, item)
	e.emitted++
	e.ctx.AllowRandom = false
	e.ctx.Rand = nil
	return nil
}

func enforceMax(val, hardCap, fallback int) int {
	if val <= 0 {
		return fallback
	}
	if val > hardCap {
		return hardCap
	}
	return val
}

func truthy(val any) bool {
	switch v := val.(type) {
	case bool:
		return v
	case nil:
		return false
	case string:
		return v != ""
	case int:
		return v != 0
	case float64:
		return v != 0
	default:
		return true
	}
}

func withTimeout(timeoutMs int, fn func() error) error {
	if timeoutMs <= 0 {
		return fn()
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()
	ch := make(chan error, 1)
	go func() {
		ch <- fn()
	}()
	select {
	case <-ctx.Done():
		return fmt.Errorf("timeout: %w", ErrTimeout)
	case err := <-ch:
		return err
	}
}
