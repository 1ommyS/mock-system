package application

import (
	"errors"

	"dsl-runner-svc/internal/dsl"
)

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrNotFound       = errors.New("not found")
	ErrConflict       = errors.New("conflict")
	ErrJobNotReady    = errors.New("job not ready")
)

const (
	ErrCodeInvalidJobPayload = "INVALID_JOB_PAYLOAD"
	ErrCodeUnauthorized      = "UNAUTHORIZED_INTERNAL"
	ErrCodeDSLParse          = "DSL_PARSE_ERROR"
	ErrCodeDSLValidation     = "DSL_VALIDATION_ERROR"
	ErrCodeDSLEval           = "DSL_EVAL_ERROR"
	ErrCodeDSLRuntime        = "DSL_RUNTIME_ERROR"
	ErrCodeDSLTimeout        = "DSL_TIMEOUT"
	ErrCodeDSLTooManyMocks   = "DSL_TOO_MANY_MOCKS"
	ErrCodeDSLResultTooLarge = "DSL_RESULT_TOO_LARGE"
	ErrCodeJobNotFound       = "JOB_NOT_FOUND"
	ErrCodeJobNotReady       = "JOB_NOT_READY"
)

type CodeError struct {
	Code string
	Err  error
}

func (e CodeError) Error() string {
	if e.Err == nil {
		return e.Code
	}
	return e.Code + ": " + e.Err.Error()
}

func (e CodeError) Unwrap() error {
	return e.Err
}

func NewCodeError(code string, err error) error {
	return CodeError{Code: code, Err: err}
}

type TransientError struct {
	Err error
}

func (e TransientError) Error() string {
	if e.Err == nil {
		return "transient error"
	}
	return e.Err.Error()
}

func (e TransientError) Unwrap() error {
	return e.Err
}

func NewTransient(err error) error {
	return TransientError{Err: err}
}

func IsTransient(err error) bool {
	var t TransientError
	return errors.As(err, &t)
}

func CodeFromError(err error) string {
	var ce CodeError
	if errors.As(err, &ce) {
		return ce.Code
	}
	switch {
	case errors.Is(err, dsl.ErrParse):
		return ErrCodeDSLParse
	case errors.Is(err, dsl.ErrValidation):
		return ErrCodeDSLValidation
	case errors.Is(err, dsl.ErrEval):
		return ErrCodeDSLEval
	case errors.Is(err, dsl.ErrTimeout):
		return ErrCodeDSLTimeout
	case errors.Is(err, dsl.ErrTooManyMocks):
		return ErrCodeDSLTooManyMocks
	case errors.Is(err, dsl.ErrResultTooLarge):
		return ErrCodeDSLResultTooLarge
	case errors.Is(err, dsl.ErrRuntime):
		return ErrCodeDSLRuntime
	}
	return ""
}
