package osoc

import (
	"log"
	"os"
)

var (
	logger *log.Logger = log.New(os.Stdout, "[service] ", log.LstdFlags|log.Lshortfile)
)
