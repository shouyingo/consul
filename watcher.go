package consul

import (
	"github.com/hashicorp/consul/api"
)

const (
	WatchAdd = iota
	WatchChange
	WatchRemove
)

type WatchFunc func(action int, id string, svc *Service)

type watchService struct {
	svc         Service
	createIndex uint64
	modifyIndex uint64
	lastIndex   uint64
}

type Watcher struct {
	c    *Consul
	name string
	tag  string
}

func (w *Watcher) Watch(wfn WatchFunc) error {
	watchsvcs := make(map[string]*watchService)
	lastIndex := uint64(0)
	client := w.c.client
	for {
		services, meta, err := client.Catalog().Service(w.name, w.tag, &api.QueryOptions{
			WaitIndex: lastIndex,
		})
		if err != nil {
			return err
		}
		lastIndex = meta.LastIndex
		for _, s := range services {
			ws := watchsvcs[s.ServiceID]
			if ws == nil {
				ws = &watchService{
					svc: Service{
						ID:      s.ServiceID,
						Name:    s.ServiceName,
						Address: s.ServiceAddress,
						Port:    s.ServicePort,
						Tags:    s.ServiceTags,
					},
					createIndex: s.CreateIndex,
					modifyIndex: s.ModifyIndex,
				}
				watchsvcs[s.ServiceID] = ws
				wfn(WatchAdd, s.ServiceID, &ws.svc)
			} else if ws.modifyIndex != s.ModifyIndex {
				ws.svc = Service{
					ID:      s.ServiceID,
					Name:    s.ServiceName,
					Address: s.ServiceAddress,
					Port:    s.ServicePort,
					Tags:    s.ServiceTags,
				}
				ws.modifyIndex = s.ModifyIndex
				wfn(WatchChange, s.ServiceID, &ws.svc)
			}
			ws.lastIndex = lastIndex
		}
		for id, ws := range watchsvcs {
			if ws.lastIndex != lastIndex {
				delete(watchsvcs, id)
				wfn(WatchRemove, id, &ws.svc)
			}
		}
	}
	return nil
}
