package application

import "errors"

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrForbidden      = errors.New("forbidden")
	ErrNotFound       = errors.New("not found")
	ErrConflict       = errors.New("conflict")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrValidation     = errors.New("validation error")
)
