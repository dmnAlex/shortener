package errx

import "errors"

var (
	ErrMethodNotAllowed = errors.New("method not allowed")
	ErrBadRequest       = errors.New("bad request")
	ErrNotFound         = errors.New("url not found")
	ErrInternalError    = errors.New("internal server error")
)
