package consul

const (
	WatchAdd = iota
	WatchChange
	WatchRemove
)

type WatchFunc func(action int, id string, svc *CatalogService)

type watchService struct {
	svc         *CatalogService
	modifyIndex uint64
	lastIndex   uint64
}

func (c *Client) WatchCatalogService(service string, tag string, wfn WatchFunc) error {
	watchsvcs := make(map[string]*watchService)
	lastIndex := uint64(0)
	for {
		services, meta, err := c.CatalogService(service, tag, &QueryOptions{
			WaitIndex: lastIndex,
		})
		if err != nil {
			if meta != nil && meta.LastIndex != lastIndex {
				lastIndex = meta.LastIndex
				continue
			}
			return err
		}
		if lastIndex == meta.LastIndex {
			continue
		}

		lastIndex = meta.LastIndex
		for i := range services {
			s := &services[i]
			ws := watchsvcs[s.ServiceID]
			if ws == nil {
				ws = &watchService{
					svc:         s,
					modifyIndex: s.ModifyIndex,
				}
				watchsvcs[s.ServiceID] = ws
				wfn(WatchAdd, s.ServiceID, ws.svc)
			} else if ws.modifyIndex != s.ModifyIndex {
				ws.svc = s
				ws.modifyIndex = s.ModifyIndex
				wfn(WatchChange, s.ServiceID, ws.svc)
			}
			ws.lastIndex = lastIndex
		}
		for id, ws := range watchsvcs {
			if ws.lastIndex != lastIndex {
				delete(watchsvcs, id)
				wfn(WatchRemove, id, ws.svc)
			}
		}
	}
	return nil
}
