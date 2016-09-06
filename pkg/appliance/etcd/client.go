package etcd

import (
	"fmt"
	"log"
	"os"
	"time"

	//"github.com/coreos/etcd/auth"
	"github.com/coreos/etcd/clientv3"

	//"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
)

var (
	logger *log.Logger = log.New(os.Stdout, "[appliance/etcd] ", log.LstdFlags|log.Lshortfile)

	etcd_endpoints       = []string{"10.3.0.212:2379"}
	etcd_dial_timeout    = 5 * time.Second
	etcd_request_timeout = 1 * time.Second
)

type V3ClientContext struct {
	endpoints      []string
	dialTimeout    time.Duration
	requestTimeout time.Duration
}

func NewV3ClientContext(endpoints []string, dialTimeout, requestTimeout time.Duration) *V3ClientContext {
	//auth.BcryptCost = bcrypt.MinCost
	c3 := &V3ClientContext{
		endpoints:      endpoints,
		dialTimeout:    dialTimeout,
		requestTimeout: requestTimeout,
	}
	if len(c3.endpoints) == 0 {
		c3.endpoints = etcd_endpoints
	}
	if int64(dialTimeout) == 0 {
		c3.dialTimeout = etcd_dial_timeout
	}
	if int64(requestTimeout) == 0 {
		c3.requestTimeout = etcd_request_timeout
	}
	return c3
}

func put(c *clientv3.Client, ctx context.Context, key, value string) (*clientv3.PutResponse, error) {
	return c.Put(ctx, key, value)
}

func (c3 *V3ClientContext) put(c *clientv3.Client, key, value string) (*clientv3.PutResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c3.requestTimeout)
	presp, err := c.Put(ctx, key, value)
	defer cancel()
	//cancel()
	if err != nil {
		logger.Printf("Failed to put k/v into Etcd v3: %s", err)
		return nil, err
	}
	return presp, nil
}

func (c3 *V3ClientContext) Put(key, value string) (*clientv3.PutResponse, error) {
	c, err := clientv3.New(clientv3.Config{
		Endpoints:   c3.endpoints,
		DialTimeout: c3.dialTimeout,
	})
	if err != nil {
		logger.Printf("Failed to setup Etcd v3 client: %s", err)
		return nil, err
	}
	defer c.Close()

	return c3.put(c, key, value)
}

func (c3 *V3ClientContext) get(c *clientv3.Client, key string) (*clientv3.GetResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c3.requestTimeout)
	gresp, err := c.Get(ctx, key)
	defer cancel()
	//cancel()
	if err != nil {
		logger.Printf("Failed to get k/v into Etcd v3: %s", err)
		return nil, err
	}
	return gresp, nil
}

func (c3 *V3ClientContext) getWithPrefix(c *clientv3.Client, prefix string) (*clientv3.GetResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c3.requestTimeout)
	defer cancel()
	// count keys about to be deleted
	gresp, err := c.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		logger.Printf("Failed to get k/v into Etcd v3: %s", err)
		return nil, err
	}
	return gresp, nil
}

func (c3 *V3ClientContext) Get(key string) (*clientv3.GetResponse, error) {
	c, err := clientv3.New(clientv3.Config{
		Endpoints:   c3.endpoints,
		DialTimeout: c3.dialTimeout,
	})
	if err != nil {
		logger.Printf("Failed to setup Etcd v3 client: %s", err)
		return nil, err
	}
	defer c.Close()

	return c3.get(c, key)
}

func (c3 *V3ClientContext) GetWithPrefix(prefix string) (*clientv3.GetResponse, error) {
	c, err := clientv3.New(clientv3.Config{
		Endpoints:   c3.endpoints,
		DialTimeout: c3.dialTimeout,
	})
	if err != nil {
		logger.Printf("Failed to setup Etcd v3 client: %s", err)
		return nil, err
	}
	defer c.Close()

	return c3.getWithPrefix(c, prefix)
}

func (c3 *V3ClientContext) delete(c *clientv3.Client, key string) (*clientv3.DeleteResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c3.requestTimeout)
	defer cancel()
	// delete the keys
	dresp, err := c.Delete(ctx, key)
	if err != nil {
		logger.Printf("Failed to delete k/v into Etcd v3: %s", err)
		return nil, err
	}

	fmt.Println("Deleted keys: " /*int64(len(gresp.Kvs)) ==*/, dresp.Deleted)

	return dresp, nil
}

func (c3 *V3ClientContext) deleteWithPrefix(c *clientv3.Client, prefix string) (*clientv3.DeleteResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c3.requestTimeout)
	defer cancel()
	// delete the keys
	dresp, err := c.Delete(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		logger.Printf("Failed to delete k/v into Etcd v3: %s", err)
		return nil, err
	}

	fmt.Println("Deleted all keys: " /*int64(len(gresp.Kvs)) ==*/, dresp.Deleted)

	return dresp, nil
}

func (c3 *V3ClientContext) Delete(key string) (*clientv3.DeleteResponse, error) {
	c, err := clientv3.New(clientv3.Config{
		Endpoints:   c3.endpoints,
		DialTimeout: c3.dialTimeout,
	})
	if err != nil {
		logger.Printf("Failed to setup Etcd v3 client: %s", err)
		return nil, err
	}
	defer c.Close()

	return c3.delete(c, key)
}

func (c3 *V3ClientContext) DeleteWithPrefix(prefix string) (*clientv3.DeleteResponse, error) {
	c, err := clientv3.New(clientv3.Config{
		Endpoints:   c3.endpoints,
		DialTimeout: c3.dialTimeout,
	})
	if err != nil {
		logger.Printf("Failed to setup Etcd v3 client: %s", err)
		return nil, err
	}
	defer c.Close()

	return c3.deleteWithPrefix(c, prefix)
}
