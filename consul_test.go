package consul

import (
	"testing"
	"time"
)

func TestServiceRegister(t *testing.T) {
	c, err := New("")
	if err != nil {
		t.Fatal(err)
	}
	id, err := c.Register(&Service{
		Name:    "consul-test",
		Address: "127.0.0.1",
		Port:    6621,
	}, time.Second*10, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Deregister(id)
	cancel := make(chan struct{})
	go func() {
		time.Sleep(time.Second * 30)
		close(cancel)
	}()
	err = c.KeepAlive(id, time.Second*9, cancel)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(11 * time.Second)
}
