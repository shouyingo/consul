package consul

import (
	"fmt"
	"io"
	"sync"
	"testing"
	"time"
)

func TestWatchCatalogService(t *testing.T) {
	c := NewClient("http://127.0.0.1:8500", "", "")
	wg := sync.WaitGroup{}
	wg.Add(2)
	go c.WatchCatalogService("consul-watch", "", func(services []CatalogService) error {
		fmt.Println("watched service:", services)
		return nil
	})
	svcmock := func(bad bool) {
		defer wg.Done()
		svc := &AgentService{
			Name:    "consul-watch",
			Address: "127.0.0.1",
			Port:    2333,
		}
		id, err := c.Register(svc, 10*time.Second, time.Minute)
		fmt.Println(">>>> register", id, bad)
		if err != nil {
			t.Fatal(err)
		}
		if bad {
			return
		}
		ch := make(chan struct{})
		time.AfterFunc(15*time.Second, func() {
			close(ch)
		})
		err = c.KeepAlive(id, 9*time.Second, ch)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(">>>> exit", id)
		c.AgentServiceDeregister(id)
	}

	go svcmock(true)
	go svcmock(false)

	wg.Wait()
}

func TestWatchKey(t *testing.T) {
	c := NewClient("http://127.0.0.1:8500", "", "")
	go func() {
		c.KVPut("test-key", []byte(`hello world`))
	}()
	c.WatchKey("test-key", func(value *KVPair) error {
		fmt.Println("watched key:", value)
		return io.EOF
	})
}
