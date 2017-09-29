package consul

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
)

type Service struct {
	ID      string
	Name    string
	Address string
	Port    int
	Tags    []string
}

type Consul struct {
	client *api.Client
}

func (c *Consul) Watcher(service string, tag string) *Watcher {
	return &Watcher{
		c:    c,
		name: service,
		tag:  tag,
	}
}

func (c *Consul) Pluse(id string) error {
	return c.client.Agent().PassTTL(id, "pluse")
}

func (c *Consul) KeepAlive(id string, ttl time.Duration, cancel <-chan struct{}) error {
	t := time.NewTicker(ttl)
	defer t.Stop()
	checkid := "service:" + id
	for {
		err := c.client.Agent().PassTTL(checkid, "keepalive")
		if err != nil {
			return err
		}
		select {
		case <-t.C:
		case <-cancel:
			return nil
		}
	}
}

func (c *Consul) Register(s *Service, ttl time.Duration, timeout time.Duration) (string, error) {
	var u [16]byte
	genuuidv4(&u)
	id := fmt.Sprintf("%s-%02x-%02x-%02x-%02x-%02x", s.Name, u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
	s.ID = id
	err := c.client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      id,
		Name:    s.Name,
		Address: s.Address,
		Port:    s.Port,
		Tags:    s.Tags,
		Check: &api.AgentServiceCheck{
			TTL: ttl.String(),
			DeregisterCriticalServiceAfter: timeout.String(),
		},
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func (c *Consul) Deregister(id string) error {
	return c.client.Agent().ServiceDeregister(id)
}

func New(addr string) (*Consul, error) {
	client, err := api.NewClient(&api.Config{
		Address: addr,
	})
	if err != nil {
		return nil, err
	}
	return &Consul{
		client: client,
	}, nil
}
