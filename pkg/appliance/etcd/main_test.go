package etcd

import (
	"fmt"
	//"log"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	//"github.com/coreos/etcd/auth"

	//"golang.org/x/crypto/bcrypt"
)

var (
	dialTimeout    = 5 * time.Second
	requestTimeout = 1 * time.Second
	endpoints      = []string{"10.3.0.212:2379", "172.17.4.50:30001"}
)

//func init() { auth.BcryptCost = bcrypt.MinCost }

func TestMain(m *testing.M) {
	if v, ok := os.LookupEnv("ETCD_V3_ADDRESSES"); ok && len(v) > 0 {
		endpoints = append(strings.Split(v, ","), endpoints...)
	}
	useCluster := true // default to running all tests
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.run=") {
			exp := strings.Split(arg, "=")[1]
			match, err := regexp.MatchString(exp, "Example")
			useCluster = (err == nil && match) || strings.Contains(exp, "Example")
			break
		}
	}

	retval := 0

	if useCluster {
		fmt.Println("Reserved for later")
	} else {
		retval = m.Run()
	}

	os.Exit(retval)
}
