package consul

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestWatch(t *testing.T) {
	c, err := New("")
	if err != nil {
		t.Fatal(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		w := c.Watcher("consul-watch", "")
		err := w.Watch(func(action int, id string, s *Service) {
			fmt.Println(">>>> onwatch", action, id, s)
		})
		if err != nil {
			t.Fatal(err)
		}
	}()
	svc := &Service{
		Name:    "consul-watch",
		Address: "127.0.0.1",
		Port:    2333,
	}
	svcmock := func(bad bool) {
		defer wg.Done()
		time.Sleep(time.Second)
		id, err := c.Register(svc, 10*time.Second, time.Minute)
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
		c.Deregister(id)
	}

	go svcmock(true)
	go svcmock(false)

	wg.Wait()
}
