package modelops

import (
	"github.com/jinzhu/copier"
	"gitlab.com/priceshield/agent-gateway/api"
)

func Clone(o *api.AgentEvent) *api.AgentEvent {
	if o == nil {
		return nil
	}

	// FIXME own impl and test perf
	c := &api.AgentEvent{}
	copier.Copy(c, o)

	//copying of slices doesn't seem to work. Seems to copy each reference instead of clone. do ourselves
	if len(o.Markets) == 0 {
		return c
	}

	//copier.Copy(c.Markets, o.Markets)
	c.Markets = make([]*api.AgentMarket, len(o.Markets))
	for i, om := range o.Markets {
		c.Markets[i] = &api.AgentMarket{}
		copier.Copy(c.Markets[i], om)

		if len(om.Options) == 0 {
			continue
		}

		c.Markets[i].Options = make([]*api.AgentOption, len(om.Options))
		for j, oo := range om.Options {
			c.Markets[i].Options[j] = &api.AgentOption{}
			copier.Copy(c.Markets[i].Options[j], oo)
		}
	}

	return c
}
