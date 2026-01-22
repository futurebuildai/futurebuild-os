package types

import "errors"

// Sentinel Errors
// L7 Standard: Typed errors for checking via errors.Is()
var (
	ErrNotFound       = errors.New("resource not found")
	ErrConflict       = errors.New("resource conflict")
	ErrUnauthorized   = errors.New("unauthorized access")
	ErrInvalidInput   = errors.New("invalid input parameters")
	ErrInternal       = errors.New("internal server error")
	ErrNotImplemented = errors.New("not implemented")
	ErrBadGateway     = errors.New("bad gateway")
)
