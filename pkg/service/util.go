package service

import (
	"errors"
)

var (
	errNotImplemented error = errors.New("Not Implemented")
	errUnexpected     error = errors.New("Unexpected")
)