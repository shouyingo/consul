// +build tested

package consul

import (
	"log"
	"testing"
)

func TestKVList(t *testing.T) {
	c := NewClient("http://127.0.0.1:8500", "", "")
	lastidx := uint64(0)
	for {
		kv, meta, err := c.KVList("k1/", &QueryOptions{WaitIndex: lastidx})
		if err != nil {
			t.Log(err)
			break
		}
		log.Printf("kv: %+v", kv)
		log.Printf("meta: %+v", meta)
		lastidx = meta.LastIndex
	}
}

func TestKVPut(t *testing.T) {
	c := NewClient("http://127.0.0.1:8500", "", "")
	t.Log(c.KVCAS("k1", []byte(`v2`), 0))
}
