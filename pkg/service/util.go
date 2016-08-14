package service

import (
	"errors"
	"log"
	"os"
)

var (
	errNotFound       error = errors.New("Not found")
	errNotImplemented error = errors.New("Not implemented")
	errUnexpected     error = errors.New("Unexpected")

	logger *log.Logger = log.New(os.Stdout, "[service package] ", log.LstdFlags|log.Lshortfile)
)
