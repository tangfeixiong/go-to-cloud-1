package utility

import (
	"log"
	"os"
)

var (
	Logger *log.Logger = log.New(os.Stdout, "[utility] ", log.LstdFlags|log.Lshortfile)
)
