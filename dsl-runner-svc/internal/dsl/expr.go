package dsl

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type EvalContext struct {
	InputBase   any
	InputParams any
	Vars        map[string]any
	LoopIndex   *int
	LoopItem    any
	Rand        *Rand
	AllowRandom bool
}

type ExprFunc func(ctx *EvalContext, args []any) (any, error)

type FuncRegistry map[string]ExprFunc

func DefaultFuncs() FuncRegistry {
	return FuncRegistry{
		"get": func(ctx *EvalContext, args []any) (any, error) {
			if len(args) < 1 || len(args) > 2 {
				return nil, fmt.Errorf("get expects 1 or 2 args")
			}
			path := args[0]
			if str, ok := path.(string); ok {
				val, err := ResolvePath(ctx, str)
				if err != nil {
					if len(args) == 2 {
						return args[1], nil
					}
					return nil, err
				}
				return val, nil
			}
			if ref, ok := path.(map[string]any); ok {
				if raw, ok := ref["$"]; ok {
					if s, ok := raw.(string); ok {
						val, err := ResolvePath(ctx, s)
						if err != nil {
							if len(args) == 2 {
								return args[1], nil
							}
							return nil, err
						}
						return val, nil
					}
				}
			}
			return nil, fmt.Errorf("get expects path string")
		},
		"concat": func(ctx *EvalContext, args []any) (any, error) {
			var b strings.Builder
			for _, arg := range args {
				b.WriteString(toString(arg))
			}
			return b.String(), nil
		},
		"join": func(ctx *EvalContext, args []any) (any, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("join expects 2 args")
			}
			arr, ok := args[0].([]any)
			if !ok {
				return nil, fmt.Errorf("join expects array")
			}
			sep := toString(args[1])
			parts := make([]string, 0, len(arr))
			for _, item := range arr {
				parts = append(parts, toString(item))
			}
			return strings.Join(parts, sep), nil
		},
		"lower": func(ctx *EvalContext, args []any) (any, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("lower expects 1 arg")
			}
			return strings.ToLower(toString(args[0])), nil
		},
		"upper": func(ctx *EvalContext, args []any) (any, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("upper expects 1 arg")
			}
			return strings.ToUpper(toString(args[0])), nil
		},
		"len": func(ctx *EvalContext, args []any) (any, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("len expects 1 arg")
			}
			switch v := args[0].(type) {
			case string:
				return len(v), nil
			case []any:
				return len(v), nil
			default:
				return nil, fmt.Errorf("len expects string or array")
			}
		},
		"toString": func(ctx *EvalContext, args []any) (any, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("toString expects 1 arg")
			}
			return toString(args[0]), nil
		},
		"toInt": func(ctx *EvalContext, args []any) (any, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("toInt expects 1 arg")
			}
			switch v := args[0].(type) {
			case float64:
				return int(v), nil
			case int:
				return v, nil
			case string:
				parsed, err := strconv.Atoi(strings.TrimSpace(v))
				if err != nil {
					return nil, err
				}
				return parsed, nil
			default:
				return nil, fmt.Errorf("toInt expects number or string")
			}
		},
		"sha256": func(ctx *EvalContext, args []any) (any, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("sha256 expects 1 arg")
			}
			return Sha256Hex([]byte(toString(args[0]))), nil
		},
		"randInt": func(ctx *EvalContext, args []any) (any, error) {
			if !ctx.AllowRandom || ctx.Rand == nil {
				return nil, fmt.Errorf("random not allowed")
			}
			if len(args) != 2 {
				return nil, fmt.Errorf("randInt expects 2 args")
			}
			min, err := toIntStrict(args[0])
			if err != nil {
				return nil, err
			}
			max, err := toIntStrict(args[1])
			if err != nil {
				return nil, err
			}
			if max < min {
				return nil, fmt.Errorf("randInt invalid range")
			}
			return ctx.Rand.Intn(max-min+1) + min, nil
		},
		"randBool": func(ctx *EvalContext, args []any) (any, error) {
			if !ctx.AllowRandom || ctx.Rand == nil {
				return nil, fmt.Errorf("random not allowed")
			}
			if len(args) != 1 {
				return nil, fmt.Errorf("randBool expects 1 arg")
			}
			p, err := toFloat(args[0])
			if err != nil {
				return nil, err
			}
			return ctx.Rand.Float64() < p, nil
		},
		"randChoice": func(ctx *EvalContext, args []any) (any, error) {
			if !ctx.AllowRandom || ctx.Rand == nil {
				return nil, fmt.Errorf("random not allowed")
			}
			if len(args) != 1 {
				return nil, fmt.Errorf("randChoice expects 1 arg")
			}
			arr, ok := args[0].([]any)
			if !ok {
				return nil, fmt.Errorf("randChoice expects array")
			}
			if len(arr) == 0 {
				return nil, fmt.Errorf("randChoice expects non-empty array")
			}
			idx := ctx.Rand.Intn(len(arr))
			return arr[idx], nil
		},
		"randString": func(ctx *EvalContext, args []any) (any, error) {
			if !ctx.AllowRandom || ctx.Rand == nil {
				return nil, fmt.Errorf("random not allowed")
			}
			if len(args) != 2 {
				return nil, fmt.Errorf("randString expects 2 args")
			}
			alpha := toString(args[0])
			length, err := toIntStrict(args[1])
			if err != nil {
				return nil, err
			}
			return ctx.Rand.String(alpha, length)
		},
		"randUUID": func(ctx *EvalContext, args []any) (any, error) {
			if !ctx.AllowRandom || ctx.Rand == nil {
				return nil, fmt.Errorf("random not allowed")
			}
			if len(args) != 0 {
				return nil, fmt.Errorf("randUUID expects 0 args")
			}
			return ctx.Rand.UUID(), nil
		},
	}
}

func EvalExpr(ctx *EvalContext, funcs FuncRegistry, expr json.RawMessage) (any, error) {
	var val any
	if err := json.Unmarshal(expr, &val); err != nil {
		return nil, err
	}
	return evalValue(ctx, funcs, val)
}

func evalValue(ctx *EvalContext, funcs FuncRegistry, val any) (any, error) {
	switch v := val.(type) {
	case map[string]any:
		if rawFn, ok := v["$fn"]; ok {
			fnName, ok := rawFn.(string)
			if !ok {
				return nil, fmt.Errorf("invalid $fn")
			}
			argsRaw, ok := v["args"]
			if !ok {
				return nil, fmt.Errorf("$fn args required")
			}
			argsArray, ok := argsRaw.([]any)
			if !ok {
				return nil, fmt.Errorf("$fn args must be array")
			}
			args := make([]any, 0, len(argsArray))
			for _, item := range argsArray {
				resolved, err := evalValue(ctx, funcs, item)
				if err != nil {
					return nil, err
				}
				args = append(args, resolved)
			}
			fn, ok := funcs[fnName]
			if !ok {
				return nil, fmt.Errorf("unknown function: %s", fnName)
			}
			return fn(ctx, args)
		}
		if rawPath, ok := v["$"]; ok && len(v) == 1 {
			path, ok := rawPath.(string)
			if !ok {
				return nil, fmt.Errorf("invalid $ path")
			}
			return ResolvePath(ctx, path)
		}
		out := make(map[string]any, len(v))
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			item := v[key]
			resolved, err := evalValue(ctx, funcs, item)
			if err != nil {
				return nil, err
			}
			out[key] = resolved
		}
		return out, nil
	case []any:
		out := make([]any, 0, len(v))
		for _, item := range v {
			resolved, err := evalValue(ctx, funcs, item)
			if err != nil {
				return nil, err
			}
			out = append(out, resolved)
		}
		return out, nil
	default:
		return v, nil
	}
}

func ResolvePath(ctx *EvalContext, path string) (any, error) {
	if path == "" {
		return nil, fmt.Errorf("empty path")
	}
	parts := parsePath(path)
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid path")
	}
	var current any
	if parts[0].key == "input" {
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid input path")
		}
		switch parts[1].key {
		case "base":
			current = ctx.InputBase
		case "params":
			current = ctx.InputParams
		default:
			return nil, fmt.Errorf("unknown input root")
		}
		parts = parts[2:]
	} else if parts[0].key == "vars" {
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid vars path")
		}
		current = ctx.Vars
		parts = parts[1:]
	} else if parts[0].key == "loop" {
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid loop path")
		}
		switch parts[1].key {
		case "index":
			if ctx.LoopIndex == nil {
				return nil, fmt.Errorf("loop.index not set")
			}
			return *ctx.LoopIndex, nil
		case "item":
			return ctx.LoopItem, nil
		default:
			return nil, fmt.Errorf("unknown loop field")
		}
	} else {
		return nil, fmt.Errorf("unknown root")
	}

	for _, part := range parts {
		switch node := current.(type) {
		case map[string]any:
			val, ok := node[part.key]
			if !ok {
				return nil, fmt.Errorf("path not found")
			}
			current = val
		case []any:
			if part.index == nil {
				return nil, fmt.Errorf("expected index")
			}
			idx := *part.index
			if idx < 0 || idx >= len(node) {
				return nil, fmt.Errorf("index out of range")
			}
			current = node[idx]
		default:
			return nil, fmt.Errorf("invalid path")
		}
	}
	return current, nil
}

type pathPart struct {
	key   string
	index *int
}

func parsePath(path string) []pathPart {
	parts := make([]pathPart, 0)
	curr := ""
	for i := 0; i < len(path); i++ {
		ch := path[i]
		switch ch {
		case '.':
			if curr != "" {
				parts = append(parts, pathPart{key: curr})
				curr = ""
			}
		case '[':
			if curr != "" {
				parts = append(parts, pathPart{key: curr})
				curr = ""
			}
			end := strings.IndexByte(path[i:], ']')
			if end == -1 {
				return nil
			}
			idxStr := path[i+1 : i+end]
			idx, err := strconv.Atoi(idxStr)
			if err != nil {
				return nil
			}
			parts = append(parts, pathPart{index: &idx})
			i += end
		default:
			curr += string(ch)
		}
	}
	if curr != "" {
		parts = append(parts, pathPart{key: curr})
	}
	return parts
}

func toString(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func toIntStrict(val any) (int, error) {
	switch v := val.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	case string:
		return strconv.Atoi(strings.TrimSpace(v))
	default:
		return 0, fmt.Errorf("expected int")
	}
}

func toFloat(val any) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(strings.TrimSpace(v), 64)
	default:
		return 0, fmt.Errorf("expected number")
	}
}
