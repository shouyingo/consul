package consul

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type QueryOptions struct {
	WaitIndex uint64
	WaitTime  time.Duration
}

type QueryMeta struct {
	LastIndex uint64
}

type CatalogService struct {
	ServiceID      string
	ServiceName    string
	ServiceAddress string
	ServicePort    int
	ServiceTags    []string

	CreateIndex uint64
	ModifyIndex uint64
}

type AgentServiceCheck struct {
	TTL                            string
	DeregisterCriticalServiceAfter string
}

type AgentService struct {
	ID      string
	Name    string
	Address string
	Port    int
	Tags    []string
	Check   AgentServiceCheck
}

type KVPair struct {
	LockIndex uint64
	Key       string
	Flags     uint64
	Value     []byte

	CreateIndex uint64
	ModifyIndex uint64
}

type request struct {
	method string
	path   string
	params []string
	body   []byte
}

func getMeta(resp *http.Response) *QueryMeta {
	meta := &QueryMeta{}
	meta.LastIndex, _ = strconv.ParseUint(resp.Header.Get("X-Consul-Index"), 10, 64)
	return meta
}

func (c *Client) newRequest(r *request) (*http.Request, error) {
	var rd io.Reader
	if r.body != nil {
		rd = bytes.NewReader(r.body)
	}
	rawurl := c.addr + r.path
	if len(r.params) > 0 {
		vals := make(url.Values, len(r.params)/2)
		for i := 0; i+1 < len(r.params); i += 2 {
			vals.Set(r.params[i], r.params[i+1])
		}
		rawurl = rawurl + "?" + vals.Encode()
	}
	req, err := http.NewRequest(r.method, rawurl, rd)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("X-Consul-Token", c.token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("%d %s", resp.StatusCode, data)
	}
	return resp, nil
}

func (c *Client) call(r *request) error {
	req, err := c.newRequest(r)
	if err == nil {
		resp, err := c.doRequest(req)
		if err == nil {
			io.CopyBuffer(ioutil.Discard, resp.Body, make([]byte, 1024))
			resp.Body.Close()
		}
	}
	return err
}

func (c *Client) query(r *request, o *QueryOptions, out interface{}) (*QueryMeta, error) {
	if o != nil {
		if o.WaitIndex != 0 {
			r.params = append(r.params, "index", strconv.FormatUint(o.WaitIndex, 10))
		}
		if o.WaitTime != 0 {
			r.params = append(r.params, "wait", strconv.FormatInt(int64(o.WaitTime/time.Millisecond), 10)+"ms")
		}
	}
	req, err := c.newRequest(r)
	if err != nil {
		return nil, err
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	meta := getMeta(resp)
	if out != nil {
		data, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return meta, json.Unmarshal(data, out)
	}
	return meta, nil
}

func (c *Client) AgentPassTTL(id string, note string) error {
	return c.call(&request{
		method: "PUT",
		path:   "/v1/agent/check/pass/" + id,
		params: []string{"note", note},
	})
}

func (c *Client) AgentServiceRegister(s *AgentService) error {
	body, _ := json.Marshal(s)
	return c.call(&request{
		method: "PUT",
		path:   "/v1/agent/service/register",
		body:   body,
	})
}

func (c *Client) AgentServiceDeregister(id string) error {
	return c.call(&request{
		method: "PUT",
		path:   "/v1/agent/service/deregister/" + id,
	})
}

func (c *Client) CatalogService(service string, tag string, options *QueryOptions) ([]CatalogService, *QueryMeta, error) {
	var svcs []CatalogService
	var params []string
	if tag != "" {
		params = []string{"tag", tag}
	}
	meta, err := c.query(&request{
		method: "GET",
		path:   "/v1/catalog/service/" + service,
		params: params,
	}, options, &svcs)
	if err != nil {
		return nil, nil, err
	}
	return svcs, meta, nil
}

func (c *Client) KVList(prefix string, options *QueryOptions) ([]KVPair, *QueryMeta, error) {
	var paris []KVPair
	meta, err := c.query(&request{
		method: "GET",
		path:   "/v1/kv/" + strings.TrimPrefix(prefix, "/"),
		params: []string{"recurse", ""},
	}, options, &paris)
	if err != nil {
		return nil, nil, err
	}
	return paris, meta, nil
}
