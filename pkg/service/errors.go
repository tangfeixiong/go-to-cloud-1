package service

import (
	"errors"
)

var (
	errBadRequest     error = errors.New("Bad request")
	errNotFound       error = errors.New("Not found")
	errNotImplemented error = errors.New("Not implemented")
	errUnexpected     error = errors.New("Unexpected")
)
