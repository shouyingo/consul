package consul

import (
	"time"
)

func (c *Client) WatchCatalogService(service string, tag string, fn func([]CatalogService, uint64) error) error {
	lastIndex := uint64(0)
	for {
		services, meta, err := c.CatalogService(service, tag, &QueryOptions{
			WaitIndex: lastIndex,
		})
		if err != nil {
			if meta != nil {
				lastIndex = meta.LastIndex
			}
			time.Sleep(time.Second)
			continue
		}
		if lastIndex == meta.LastIndex {
			continue
		}
		lastIndex = meta.LastIndex
		err = fn(services, lastIndex)
		if err != nil {
			return err
		}
	}
}

func (c *Client) WatchKey(key string, fn func(*KVPair, uint64) error) error {
	lastIndex := uint64(0)
	for {
		value, meta, err := c.KVGet(key, &QueryOptions{
			WaitIndex: lastIndex,
		})
		if err != nil {
			if meta != nil {
				lastIndex = meta.LastIndex
			}
			time.Sleep(time.Second)
			continue
		}
		if meta.LastIndex == lastIndex {
			continue
		}
		lastIndex = meta.LastIndex
		err = fn(value, lastIndex)
		if err != nil {
			return err
		}
	}
}
