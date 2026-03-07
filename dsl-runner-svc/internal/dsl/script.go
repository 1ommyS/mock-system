package dsl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type Script struct {
	Version int    `json:"version"`
	Steps   []Step `json:"steps"`
}

type Step interface {
	Op() string
}

type LetStep struct {
	OpName string          `json:"op"`
	Name   string          `json:"name"`
	Value  json.RawMessage `json:"value"`
}

func (s LetStep) Op() string { return s.OpName }

type IfStep struct {
	OpName string          `json:"op"`
	Cond   json.RawMessage `json:"cond"`
	Then   []Step          `json:"then"`
	Else   []Step          `json:"else"`
}

func (s IfStep) Op() string { return s.OpName }

type RepeatStep struct {
	OpName string          `json:"op"`
	Count  json.RawMessage `json:"count"`
	Do     []Step          `json:"do"`
}

func (s RepeatStep) Op() string { return s.OpName }

type ForEachStep struct {
	OpName string          `json:"op"`
	Items  json.RawMessage `json:"items"`
	Do     []Step          `json:"do"`
}

func (s ForEachStep) Op() string { return s.OpName }

type EmitStep struct {
	OpName           string          `json:"op"`
	Name             json.RawMessage `json:"name"`
	RequestMatch     json.RawMessage `json:"requestMatch"`
	ResponseTemplate json.RawMessage `json:"responseTemplate"`
	Meta             json.RawMessage `json:"meta"`
}

func (s EmitStep) Op() string { return s.OpName }

type AssertStep struct {
	OpName string          `json:"op"`
	Cond   json.RawMessage `json:"cond"`
	Error  string          `json:"error"`
}

func (s AssertStep) Op() string { return s.OpName }

func ParseScript(data []byte) (Script, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return Script{}, fmt.Errorf("empty script")
	}
	if trimmed[0] != '{' {
		script, err := parseTextScript(trimmed)
		if err != nil {
			return Script{}, fmt.Errorf("parse script: %w", err)
		}
		return script, nil
	}
	script, err := parseJSONScript(trimmed)
	if err != nil {
		return Script{}, fmt.Errorf("parse script: %w", err)
	}
	return script, nil
}

func parseJSONScript(data []byte) (Script, error) {
	var raw struct {
		Version int               `json:"version"`
		Steps   []json.RawMessage `json:"steps"`
	}
	if err := decodeStrict(data, &raw); err != nil {
		return Script{}, fmt.Errorf("parse script: %w", err)
	}
	if raw.Version != 1 {
		return Script{}, fmt.Errorf("unsupported version: %d", raw.Version)
	}
	if len(raw.Steps) == 0 {
		return Script{}, fmt.Errorf("steps are required")
	}
	steps, err := parseSteps(raw.Steps)
	if err != nil {
		return Script{}, err
	}
	return Script{Version: raw.Version, Steps: steps}, nil
}

func parseSteps(raw []json.RawMessage) ([]Step, error) {
	steps := make([]Step, 0, len(raw))
	for i, item := range raw {
		step, err := parseStep(item)
		if err != nil {
			return nil, fmt.Errorf("step %d: %w", i, err)
		}
		steps = append(steps, step)
	}
	return steps, nil
}

func parseStep(data json.RawMessage) (Step, error) {
	var base struct {
		Op string `json:"op"`
	}
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, fmt.Errorf("parse op: %w", err)
	}
	switch base.Op {
	case "let":
		var step LetStep
		if err := decodeStrict(data, &step); err != nil {
			return nil, fmt.Errorf("parse let: %w", err)
		}
		if step.Name == "" || len(step.Value) == 0 {
			return nil, fmt.Errorf("let requires name and value")
		}
		return step, nil
	case "if":
		var raw struct {
			Op   string            `json:"op"`
			Cond json.RawMessage   `json:"cond"`
			Then []json.RawMessage `json:"then"`
			Else []json.RawMessage `json:"else"`
		}
		if err := decodeStrict(data, &raw); err != nil {
			return nil, fmt.Errorf("parse if: %w", err)
		}
		if len(raw.Cond) == 0 {
			return nil, fmt.Errorf("if requires cond")
		}
		if raw.Then == nil {
			return nil, fmt.Errorf("if requires then")
		}
		thenSteps, err := parseSteps(raw.Then)
		if err != nil {
			return nil, err
		}
		elseSteps := []Step{}
		if raw.Else != nil {
			elseSteps, err = parseSteps(raw.Else)
			if err != nil {
				return nil, err
			}
		}
		return IfStep{OpName: raw.Op, Cond: raw.Cond, Then: thenSteps, Else: elseSteps}, nil
	case "repeat":
		var raw struct {
			Op    string            `json:"op"`
			Count json.RawMessage   `json:"count"`
			Do    []json.RawMessage `json:"do"`
		}
		if err := decodeStrict(data, &raw); err != nil {
			return nil, fmt.Errorf("parse repeat: %w", err)
		}
		if len(raw.Count) == 0 {
			return nil, fmt.Errorf("repeat requires count")
		}
		if raw.Do == nil {
			return nil, fmt.Errorf("repeat requires do")
		}
		doSteps, err := parseSteps(raw.Do)
		if err != nil {
			return nil, err
		}
		return RepeatStep{OpName: raw.Op, Count: raw.Count, Do: doSteps}, nil
	case "forEach":
		var raw struct {
			Op    string            `json:"op"`
			Items json.RawMessage   `json:"items"`
			Do    []json.RawMessage `json:"do"`
		}
		if err := decodeStrict(data, &raw); err != nil {
			return nil, fmt.Errorf("parse forEach: %w", err)
		}
		if len(raw.Items) == 0 {
			return nil, fmt.Errorf("forEach requires items")
		}
		if raw.Do == nil {
			return nil, fmt.Errorf("forEach requires do")
		}
		doSteps, err := parseSteps(raw.Do)
		if err != nil {
			return nil, err
		}
		return ForEachStep{OpName: raw.Op, Items: raw.Items, Do: doSteps}, nil
	case "emit":
		var step EmitStep
		if err := decodeStrict(data, &step); err != nil {
			return nil, fmt.Errorf("parse emit: %w", err)
		}
		if len(step.Name) == 0 || len(step.RequestMatch) == 0 || len(step.ResponseTemplate) == 0 || len(step.Meta) == 0 {
			return nil, fmt.Errorf("emit requires name, requestMatch, responseTemplate, meta")
		}
		return step, nil
	case "assert":
		var step AssertStep
		if err := decodeStrict(data, &step); err != nil {
			return nil, fmt.Errorf("parse assert: %w", err)
		}
		if len(step.Cond) == 0 || step.Error == "" {
			return nil, fmt.Errorf("assert requires cond and error")
		}
		return step, nil
	default:
		return nil, fmt.Errorf("unknown op: %s", base.Op)
	}
}

func decodeStrict(data []byte, dst any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("unexpected trailing data")
	}
	return nil
}
