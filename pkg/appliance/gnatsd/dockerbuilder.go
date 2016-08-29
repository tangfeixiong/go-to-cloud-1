package gnatsd

import (
	"log"
	"os"

	"github.com/cloudfoundry/yagnats"
)

func Publish(addrs []string, user, password *string, subject string, message []byte) {
	var logger *log.Logger = log.New(os.Stdout, "[appliance/gnatsd, Publish] ", log.LstdFlags|log.Lshortfile)

	if len(subject) > 0 {
		_subject = subject
	}
	if len(message) > 0 {
		_message = message
	}

	c, err := ClientWithConnection(addrs, user, password)
	if err != nil {
		logger.Printf("Wrong auth or something failed to contact gnatsd: %+v", err)
		return
	}

	if err := c.Publish(_subject, _message); err != nil {
		logger.Printf("Faile to publish into gnatsd: %+v", err)
	}
	c.Disconnect()
}

func Subscribe(addrs []string, user, password *string, subject string) ([]byte, error) {
	var logger *log.Logger = log.New(os.Stdout, "[appliance/gnatsd, Publish] ", log.LstdFlags|log.Lshortfile)

	if len(subject) > 0 {
		_subject = subject
	}

	c, err := ClientWithConnection(addrs, user, password)
	if err != nil {
		logger.Printf("Wrong auth or something failed to contact gnatsd: %+v", err)
		return []byte{}, err
	}

	var data []byte
	id, err := c.Subscribe(_subject, func(msg *yagnats.Message) {
		data = msg.Payload
	})
	if err != nil {
		logger.Printf("Faile to subscribe into gnatsd: %+v", err)
		return []byte{}, err
	}
	logger.Printf("Got message(id=%d): %s\n", id, string(data))
	c.Disconnect()
	return data, nil
}
