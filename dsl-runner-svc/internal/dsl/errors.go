package dsl

import "errors"

var (
	ErrParse          = errors.New("dsl parse error")
	ErrValidation     = errors.New("dsl validation error")
	ErrEval           = errors.New("dsl eval error")
	ErrRuntime        = errors.New("dsl runtime error")
	ErrTimeout        = errors.New("dsl timeout")
	ErrTooManyMocks   = errors.New("dsl too many mocks")
	ErrResultTooLarge = errors.New("dsl result too large")
)
