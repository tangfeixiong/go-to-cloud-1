package service

import (
	"errors"
)

var (
	errBadRequest     error = errors.New("Bad request")
	errNotFound       error = errors.New("Not found")
	errNotImplemented error = errors.New("Not implemented")
	errNotSupported   error = errors.New("Not supported")
	errUnexpected     error = errors.New("Unexpected")
)
