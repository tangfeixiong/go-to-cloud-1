package gnatsd

import (
	"os"
	"strings"

	"github.com/cloudfoundry/yagnats"
)

var (
	_addrs    []string = []string{"10.3.0.39:4222"}
	_username string   = "derek"
	_password string   = "T0pS3cr3t"
	_subject  string   = "hello"
	_message  []byte   = []byte("world")
)

func init() {
	if v, ok := os.LookupEnv("GNATSD_ADDRESSES"); ok && len(v) > 0 {
		_addrs = strings.Split(v, ",")
	}
}

func ClientWithConnection(addrs []string, user, password *string) (*yagnats.Client, error) {
	client := yagnats.NewClient()

	if len(addrs) > 0 {
		_addrs = append(addrs, _addrs...)
	}
	if user != nil {
		_username = *user
	}
	if password != nil {
		_password = *password
	}

	err := client.Connect(&yagnats.ConnectionInfo{
		Addr:     _addrs[0],
		Username: _username,
		Password: _password,
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}
