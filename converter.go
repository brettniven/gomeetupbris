package gomeetupbris

import (
	"strconv"

	"github.com/golang/protobuf/ptypes/wrappers"
	"gitlab.com/priceshield/agent-gateway/api"
)

// converter converts the provider format to the common format
type converter struct {
	in       <-chan pipelineItem // with provider state
	out      chan<- pipelineItem // with converted state
	doneChan chan bool
}

func newConverter(in <-chan pipelineItem, out chan<- pipelineItem) *converter {
	return &converter{
		in:       in,
		out:      out,
		doneChan: make(chan bool),
	}
}

func (c *converter) start() {
	for {
		select {
		case pi := <-c.in:
			go func(pi pipelineItem) {
				converted := convertEvent(pi.providerEvent)
				pi.converted = converted
				// forward on to the next step in the pipeline
				c.out <- pi
			}(pi)
		case <-c.doneChan:
			return
		}
	}
}

func (c *converter) stop() {
	c.doneChan <- true
}

func convertEvent(pe ProviderEvent) api.AgentEvent {

	e := api.AgentEvent{
		ProviderID:  strconv.FormatInt(pe.MasterEventID, 10),
		Name:        &wrappers.StringValue{Value: pe.MasterEventName},
		Sport:       &wrappers.StringValue{Value: pe.Sport},
		Competition: &wrappers.StringValue{Value: pe.DisplayMasterCategoryName},
	}

	m := convertMarket(pe.FeaturedMarket)
	e.Markets = []*api.AgentMarket{&m}

	return e
}

func convertMarket(pm FeaturedMarket) api.AgentMarket {

	marketName := standardizeMarketName(pm.EventName) // oddly named, but this is the market name

	m := api.AgentMarket{
		ProviderID: strconv.FormatInt(pm.EventID, 10),
		Name:       &wrappers.StringValue{Value: marketName},
		Hidden:     &wrappers.BoolValue{Value: false}, // if its there, it's visible
		Suspended:  &wrappers.BoolValue{Value: !pm.IsOpenForBetting},
		Options:    make([]*api.AgentOption, 0),
	}

	for _, po := range pm.Outcomes {
		o := convertOption(po)
		m.Options = append(m.Options, &o)
	}

	return m
}

func convertOption(po Outcome) api.AgentOption {
	o := api.AgentOption{
		ProviderID: strconv.FormatInt(po.OutcomeID, 10),
		Name:       &wrappers.StringValue{Value: po.OutcomeName},
		Hidden:     &wrappers.BoolValue{Value: false}, // if its there, it's visible
		Suspended:  &wrappers.BoolValue{Value: false}, // hmm, not sure if we can get a status for this...
		Price:      &wrappers.DoubleValue{Value: po.Price},
	}

	return o
}
