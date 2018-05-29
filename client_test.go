package consul

import (
	"os"
	"strconv"
	"testing"
	"time"
)

func TestServiceRegister(t *testing.T) {
	c := NewClient("http://127.0.0.1:8500", "", "")
	id, err := c.Register(&AgentService{
		Name:    "consul-test",
		Address: "127.0.0.1",
		Port:    6621,
		Meta:    map[string]string{"pid": strconv.Itoa(os.Getpid())},
		Tags:    []string{"v1"},
	}, time.Second*10, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer c.AgentServiceDeregister(id)
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
