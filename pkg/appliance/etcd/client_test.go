package etcd

import (
	"fmt"
	"log"
	"testing"

	"github.com/coreos/etcd/clientv3"

	"golang.org/x/net/context"
)

func TestKV_withprefix(t *testing.T) {
	cli := NewV3ClientContext([]string{}, 0, 0)

	v := "v1"
	c := "default"
	key1 := fmt.Sprintf("/apaasapis/%s/clusters/%s", v, c)
	ns := "default"
	buildconfigname := "osobuilds"
	buildname := "osobuilds"
	key2 := fmt.Sprintf("/apaasapis/%s/clusters/%s/projects/%s/builders/%s/builds/%s", v, c, ns, buildconfigname, buildname)

	presp, err := cli.Put(key2, "foo")
	if err != nil {
		log.Fatal(err)
	}
	t.Log(presp)

	rresp, err := cli.Get(key2)
	if err != nil {
		log.Fatal(err)
	}
	for _, ev := range rresp.Kvs {
		t.Logf("%s : %s\n", ev.Key, ev.Value)
	}

	rresp, err = cli.GetWithPrefix(key1)
	if err != nil {
		log.Fatal(err)
	}
	for _, ev := range rresp.Kvs {
		t.Logf("%s : %s\n", ev.Key, ev.Value)
	}

	rresp, err = cli.GetWithPrefix(key2)
	if err != nil {
		log.Fatal(err)
	}
	for _, ev := range rresp.Kvs {
		t.Logf("%s : %s\n", ev.Key, ev.Value)
	}
}

func TestKV_fromprefix(t *testing.T) {
	cli := NewV3ClientContext([]string{}, 0, 0)

	v := "v1"
	c := "default"
	key := fmt.Sprintf("apaasapis%sclusters%", v, c)
	//ns := "default"
	//buildconfigname := "osobuilds"
	//buildname := "osobuilds"
	//key := fmt.Sprintf("apaasapis%sclusters%projects%sbuilders%sbuilds%s", v, c, ns, buildconfigname, buildname)

	resp, err := cli.GetWithPrefix(key)
	if err != nil {
		log.Fatal(err)
	}
	for _, ev := range resp.Kvs {
		t.Logf("%s : %s\n", ev.Key, ev.Value)
	}
}

func TestKV_put(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcd_endpoints[:1],
		DialTimeout: etcd_dial_timeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), etcd_request_timeout)
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
		Endpoints:   etcd_endpoints[:1],
		DialTimeout: etcd_dial_timeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	_, err = cli.Put(context.TODO(), "foo", "bar")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), etcd_request_timeout)
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
