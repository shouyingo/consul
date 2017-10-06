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

func readError(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return fmt.Errorf("%d %s", resp.StatusCode, data)
	}
	return nil
}

func discardBody(resp *http.Response) {
	io.CopyBuffer(ioutil.Discard, resp.Body, make([]byte, 1024))
	resp.Body.Close()
}

func getMeta(resp *http.Response) *QueryMeta {
	meta := &QueryMeta{}
	meta.LastIndex, _ = strconv.ParseUint(resp.Header.Get("X-Consul-Index"), 10, 64)
	return meta
}

func (c *Client) doRequest(r *request) (*http.Response, error) {
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
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("X-Consul-Token", c.token)
	}
	return http.DefaultClient.Do(req)
}

func (c *Client) call(r *request) error {
	resp, err := c.doRequest(r)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return readError(resp)
	}
	discardBody(resp)
	return nil
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
	resp, err := c.doRequest(r)
	if err != nil {
		return nil, err
	}

	meta := getMeta(resp)
	if resp.StatusCode == http.StatusOK {
		if out != nil {
			data, _ := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			err = json.Unmarshal(data, out)
		} else {
			discardBody(resp)
		}
	} else {
		err = readError(resp)
	}
	return meta, err
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
	var pairs []KVPair
	meta, err := c.query(&request{
		method: "GET",
		path:   "/v1/kv/" + strings.TrimPrefix(prefix, "/"),
		params: []string{"recurse", ""},
	}, options, &pairs)
	return pairs, meta, err
}

func (c *Client) KVPut(key string, value string) error {
	return c.call(&request{
		method: "PUT",
		path:   "/v1/kv/" + strings.TrimPrefix(key, "/"),
		body:   []byte(value),
	})
}
