package gnatsd

import (
	"fmt"
	"testing"

	"github.com/cloudfoundry/yagnats"
)

func TestYagnats_simple(t *testing.T) {
	client := yagnats.NewClient()

	err := client.Connect(&yagnats.ConnectionInfo{
		Addr:     "10.3.0.39:4222",
		Username: "derek",
		Password: "T0pS3cr3t",
	})
	if err != nil {
		t.Fatal(err)
		//panic("Wrong auth or something.")
	}

	var data yagnats.Message
	client.Subscribe("some.subject", func(msg *yagnats.Message) {
		data = *msg
		fmt.Printf("Got message: %s\n", msg.Payload)
	})

	client.Publish("some.subject", []byte("Sup son?"))

	client.Publish("some.subject", []byte("Again?"))

	t.Log(string(data.Payload))
}
