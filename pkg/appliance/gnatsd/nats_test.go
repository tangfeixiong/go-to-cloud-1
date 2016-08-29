package gnatsd

import (
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/nats"
)

var (
	timeout = 60 * time.Second // nats.DefaultTimeout
)

func TestNats_simple(t *testing.T) {
	optfunc := func(opts *nats.Options) error {
		opts.User = "derek"
		opts.Password = "T0pS3cr3t"
		return nil
	}
	nc, err := nats.Connect("nats://10.3.0.39:4222", optfunc)
	if err != nil {
		t.Fatal(err)
	}

	var data nats.Msg
	// Simple Async Subscriber
	nc.Subscribe("foo", func(m *nats.Msg) {
		data = *m
		fmt.Printf("Received a message: %s\n", string(m.Data))
	})
	fmt.Printf("Message: %+v", data)

	nc.Publish("foo", []byte("Hello World"))

	// Channel Subscriber
	//	ch := make(chan *nats.Msg, 64)
	//	sub, err := nc.ChanSubscribe("foo", ch)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	msg := <-ch
	//	if len(msg.Data) > 0 {
	//		fmt.Printf("Received a message: %s\n", string(msg.Data))
	//	} else {
	//		fmt.Printf("Message: %+v", msg)
	//	}
	//	sub.Unsubscribe()

	// Simple Sync Subscriber
	//	sub, err := nc.SubscribeSync("foo")
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	m, err := sub.NextMsg(timeout)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	fmt.Println(string(m.Data))
	//	t.Logf("Message: %+v", m)

	nc.Close()
}
