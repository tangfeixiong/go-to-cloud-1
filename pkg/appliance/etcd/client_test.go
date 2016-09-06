package etcd

import (
	"fmt"
	"log"
	"testing"

	"github.com/coreos/etcd/clientv3"

	"golang.org/x/net/context"
)

func TestKV_put(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints[:1],
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	_, err = cli.Put(ctx, "sample_key", "sample_value")
	defer cancel()
	//cancel()
	if err != nil {
		log.Fatal(err)
	}

	// count keys about to be deleted
	gresp, err := cli.Get(ctx, "sample_key", clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
	}

	// delete the keys
	dresp, err := cli.Delete(ctx, "sample_key", clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Deleted all keys:", int64(len(gresp.Kvs)) == dresp.Deleted)
	// Output:
	t.Log("// Deleted all keys: true")
}

func TestKV_get(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints[:1],
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	_, err = cli.Put(context.TODO(), "foo", "bar")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := cli.Get(ctx, "foo")
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s : %s\n", ev.Key, ev.Value)
	}

	t.Log("// Output: foo : bar")
}
