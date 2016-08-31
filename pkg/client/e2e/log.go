package e2e

import (
	"log"
	"os"
)

var (
	logger *log.Logger = log.New(os.Stdout, "[client/e2e] ", log.LstdFlags|log.Lshortfile)
)
