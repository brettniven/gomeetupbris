package gomeetupbris

import (
	gocache "github.com/patrickmn/go-cache"
	"gitlab.com/priceshield/agent-gateway/api"
	"time"
)

// cache is just a simple wrapper around go-cache, to use concrete types
type cache struct {
	eventCache *gocache.Cache // key is event provider id, value is our full known state of the event (api.AgentEvent)
}

func newCache() *cache {
	return &cache{
		eventCache: gocache.New(1*time.Hour, 10*time.Minute),
	}
}

func (c *cache) get(providerEventID string) (api.AgentEvent, bool) {
	cacheObj, ok := c.eventCache.Get(providerEventID)
	if ok {
		initialFullState := cacheObj.(api.AgentEvent)
		return initialFullState, true
	}

	return api.AgentEvent{}, false
}

func (c *cache) set(ae api.AgentEvent) {
	c.eventCache.Set(ae.ProviderID, ae, gocache.DefaultExpiration)
}
