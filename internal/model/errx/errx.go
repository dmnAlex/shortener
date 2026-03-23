package errx

import "errors"

var (
	ErrAlreadyExists    = errors.New("already exists")
	ErrMethodNotAllowed = errors.New("method not allowed")
	ErrBadRequest       = errors.New("bad request")
	ErrNotFound         = errors.New("not found")
	ErrConflict         = errors.New("conflict")
	ErrUnprocessable    = errors.New("unprocessable")
	ErrInternalError    = errors.New("internal server error")
	ErrGone             = errors.New("gone")
	ErrUnauthorized     = errors.New("unauthorized")
)
