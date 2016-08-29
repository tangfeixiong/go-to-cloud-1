package service

import (
	"log"
	"os"
)

var (
	logger *log.Logger = log.New(os.Stdout, "[pkg/service] ", log.LstdFlags|log.Lshortfile)
)
