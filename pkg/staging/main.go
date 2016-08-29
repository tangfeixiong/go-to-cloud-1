package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var runfunc = run

func main() {
	go runfunc()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

	// Block until a signal is received.
	<-c

}

func run() {
	go func() {
		time.Sleep(5 * time.Second)
		fmt.Println("concurrency end")
	}()

	fmt.Println("primary end")

}
