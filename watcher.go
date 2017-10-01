package consul

const (
	WatchAdd = iota
	WatchChange
	WatchRemove
)

type WatchFunc func(action int, id string, svc *AgentService)

type watchService struct {
	svc         AgentService
	createIndex uint64
	modifyIndex uint64
	lastIndex   uint64
}

type Watcher struct {
	c    *Client
	name string
	tag  string
}

func (w *Watcher) Watch(wfn WatchFunc) error {
	watchsvcs := make(map[string]*watchService)
	lastIndex := uint64(0)
	for {
		services, meta, err := w.c.CatalogService(w.name, w.tag, &QueryOptions{
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
					svc: AgentService{
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
				ws.svc = AgentService{
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
