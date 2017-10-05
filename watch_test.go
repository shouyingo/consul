// +build tested

package consul

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestWatch(t *testing.T) {
	c := New("http://127.0.0.1:8500", "")
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		err := c.Watch("consul-watch", "", func(action int, id string, s *CatalogService) {
			fmt.Println(">>>> onwatch", action, id, s)
		})
		if err != nil {
			t.Fatal(err)
		}
	}()
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
		time.AfterFunc(20*time.Second, func() {
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
