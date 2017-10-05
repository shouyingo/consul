// +build tested

package consul

import (
	"log"
	"testing"
)

func TestKVList(t *testing.T) {
	c := New("http://127.0.0.1:8500", "")
	lastidx := uint64(0)
	for {
		kv, meta, err := c.KVList("k1/", &QueryOptions{WaitIndex: lastidx})
		if err != nil {
			t.Fatal(err)
		}
		log.Printf("kv: %+v", kv)
		log.Printf("meta: %+v", meta)
		lastidx = meta.LastIndex
	}
}
