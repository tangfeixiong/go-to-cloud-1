package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/client/e2e"
)

func main() {
	var status int
	var ok bool

	status, ok = e2e.DockerBuildWithNewConfig()
	if !ok || status == 0 {
		fmt.Println("create error")
		os.Exit(1)
	}
	fmt.Printf("Status to creation call: %+v\n", status)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			time.Sleep(1 * time.Second)
			status, ok = e2e.TrackingDockerBuild()
			if !ok || status == 0 {
				fmt.Println("Received failed")
				return
			}
			switch status {
			case 1:
				fmt.Printf("Continuning: %+v\n", status)
			case 2:
				fmt.Printf("Failure: %+v\n", status)
				return
			case 3:
				fmt.Printf("Succeeded: %+v\n", status)
				return
			case 4:
				fmt.Printf("Warning: %+v\n", status)
				return
			default:
				fmt.Println("Unexpected")
				return
			}
		}
	}()
	wg.Wait()

	status, ok = e2e.FindDockerBuild()
	if !ok || status == 0 {
		fmt.Println("find error")
		os.Exit(1)
	}

	fmt.Printf("Status to retrieve call: %+v\n", status)
}
