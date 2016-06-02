package logger

import (
	"log"
	"os"
)

var (
	Logger *log.Logger = log.New(os.Stdout, "[tangfx] ", log.LstdFlags|log.Lshortfile)
)
