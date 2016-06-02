package main

import (
	"log"
	"os"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/openshift/client"
)

func main() {
	//	user, err := client.WhoAmI()
	//	if err != nil {
	//		log.Printf("\nCould not know user: %s", err)
	//		os.Exit(1)
	//	}
	//	log.Printf("\nuser: %v", user)

	//	users, err := client.Whole()
	//	if err != nil {
	//		log.Printf("\nCould not know user: %s", err)
	//		os.Exit(1)
	//	}
	//	log.Printf("\nuser: %v", users)

	//	if err := client.LoginWithBasicAuth("tangfeixiong", "tangfeixiong"); err != nil {
	//		log.Fatalln(err)
	//	}
	//	os.Exit(0)

	if err := client.ShowUsers(); err != nil {
		log.Fatal(err)
	}

	if err := client.ShowSelf(); err != nil {
		log.Fatal(err)
	}

	if err := client.ShowProjects(); err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
