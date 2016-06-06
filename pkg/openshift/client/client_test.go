package client

import (
	"log"
	"os"

	"testing"
)

func TestClient(t *testing.T) {
	//	user, err := WhoAmI()
	//	if err != nil {
	//		log.Printf("\nCould not know user: %s", err)
	//		os.Exit(1)
	//	}
	//	log.Printf("\nuser: %v", user)

	//	users, err := Whole()
	//	if err != nil {
	//		log.Printf("\nCould not know user: %s", err)
	//		os.Exit(1)
	//	}
	//	log.Printf("\nuser: %v", users)

	//	if err := LoginWithBasicAuth("tangfeixiong", "tangfeixiong"); err != nil {
	//		log.Fatalln(err)
	//	}

	if err := DoBasicAuth(); err != nil {
		log.Fatal(err)
	}

	//	if err := ShowUsers(); err != nil {
	//		log.Fatal(err)
	//	}

	//	if err := ShowSelf(); err != nil {
	//		log.Fatal(err)
	//	}

	//	if err := ShowProjects(); err != nil {
	//		log.Fatal(err)
	//	}

	os.Exit(0)
}
