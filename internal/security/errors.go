package security

import "errors"

var (
	ErrPasswordNotSet    = errors.New("password not set")
	ErrPasswordCorrupted = errors.New("password data corrupted")
	ErrInvalidInput      = errors.New("invalid input")
)
