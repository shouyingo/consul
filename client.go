package consul

import (
	"fmt"
	"strings"
	"time"
)

type Client struct {
	addr  string
	token string
	dc    string
}

func (c *Client) KeepAlive(id string, ttl time.Duration, cancel <-chan struct{}) error {
	t := time.NewTicker(ttl)
	defer t.Stop()
	checkid := "service:" + id
	for {
		err := c.AgentPassTTL(checkid, "keepalive")
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

func (c *Client) Register(s *AgentService, ttl time.Duration, timeout time.Duration) (string, error) {
	if s.ID == "" {
		var u [16]byte
		genuuidv4(&u)
		id := fmt.Sprintf("%s-%02x-%02x-%02x-%02x-%02x", s.Name, u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
		s.ID = id
	}
	s.Check.TTL = ttl.String()
	s.Check.DeregisterCriticalServiceAfter = timeout.String()
	err := c.AgentServiceRegister(s)
	if err != nil {
		return "", err
	}
	return s.ID, nil
}

func New(addr string, token string) *Client {
	addr = strings.Trim(addr, "/")
	if !strings.Contains(addr, "://") {
		addr = "http://" + addr
	}
	return &Client{
		addr:  addr,
		token: token,
	}
}

func NewClient(addr string, token string, dc string) *Client {
	addr = strings.Trim(addr, "/")
	if !strings.Contains(addr, "://") {
		addr = "http://" + addr
	}
	return &Client{
		addr:  addr,
		token: token,
		dc:    dc,
	}
}
