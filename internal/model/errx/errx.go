package errx

import "errors"

var (
	ErrMethodNotAllowed = errors.New("method not allowed")
	ErrBadRequest       = errors.New("bad request")
	ErrNotFound         = errors.New("not found")
	ErrConflict         = errors.New("conflict")
	ErrInternalError    = errors.New("internal server error")
)
