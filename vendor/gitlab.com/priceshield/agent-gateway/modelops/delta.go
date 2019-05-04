package modelops

import (
	"errors"
	"github.com/golang/protobuf/ptypes/wrappers"

	"gitlab.com/priceshield/agent-gateway/api"
)

// Delta computes the difference between a and b, with a as the base
func Delta(a, b *api.AgentEvent) (*api.AgentEvent, error) {

	if b == nil {
		return nil, nil
	}

	if a == nil {
		return &*b, nil
	}

	// if we are here, both are non-nil
	if a.ProviderID != b.ProviderID {
		return nil, errors.New("events_with_differing_ids")
	}

	delta := &api.AgentEvent{}

	delta.Name = stringValueDelta(a.Name, b.Name)
	delta.Sport = stringValueDelta(a.Sport, b.Sport)
	delta.Competition = stringValueDelta(a.Competition, b.Competition)
	delta.Hidden = boolValueDelta(a.Hidden, b.Hidden)
	delta.Suspended = boolValueDelta(a.Suspended, b.Suspended)

	delta.Markets = marketsDelta(a.Markets, b.Markets)

	// return nil if there was no delta
	if delta.Name == nil && delta.Sport == nil && delta.Competition == nil &&
		delta.Hidden == nil && delta.Suspended == nil && delta.Markets == nil {
		return nil, nil
	}

	// we're non-nil. Add the event id
	delta.ProviderID = a.ProviderID

	return delta, nil
}

func marketsDelta(a, b []*api.AgentMarket) []*api.AgentMarket {
	if b == nil {
		return nil
	}

	if a == nil {
		return b
	}

	// delta elements
	r := make([]*api.AgentMarket, 0)
	for _, bm := range b {

		am := findMatchingMarket(a, bm)

		if am == nil {
			// exists in b but not a. copy
			r = append(r, bm)
			continue
		}

		// exists in both
		dm := marketDelta(am, bm)
		if dm != nil {
			r = append(r, dm)
		}
	}

	if len(r) == 0 {
		return nil
	}

	return r
}

func marketDelta(a, b *api.AgentMarket) *api.AgentMarket {

	if b == nil {
		return nil
	}

	if a == nil {
		return &*b
	}

	delta := &api.AgentMarket{}

	delta.Name = stringValueDelta(a.Name, b.Name)
	delta.Hidden = boolValueDelta(a.Hidden, b.Hidden)
	delta.Suspended = boolValueDelta(a.Suspended, b.Suspended)

	delta.Options = optionsDelta(a.Options, b.Options)

	// return nil if there was no delta
	if delta.Name == nil && delta.Hidden == nil && delta.Suspended == nil && delta.Options == nil {
		return nil
	}

	// we're non-nil. Add the provider id
	delta.ProviderID = a.ProviderID

	return delta
}

func optionsDelta(a, b []*api.AgentOption) []*api.AgentOption {
	if b == nil {
		return nil
	}

	if a == nil {
		return b
	}

	// delta elements
	r := make([]*api.AgentOption, 0)
	for _, bo := range b {

		ao := findMatchingOption(a, bo)

		if ao == nil {
			// exists in b but not a. copy
			r = append(r, bo)
			continue
		}

		// exists in both
		dm := optionDelta(ao, bo)
		if dm != nil {
			r = append(r, dm)
		}
	}

	if len(r) == 0 {
		return nil
	}

	return r
}

func optionDelta(a, b *api.AgentOption) *api.AgentOption {

	if b == nil {
		return nil
	}

	if a == nil {
		return &*b
	}

	delta := &api.AgentOption{}

	delta.Name = stringValueDelta(a.Name, b.Name)
	delta.Price = doubleValueDelta(a.Price, b.Price)
	delta.Hidden = boolValueDelta(a.Hidden, b.Hidden)
	delta.Suspended = boolValueDelta(a.Suspended, b.Suspended)

	// return nil if there was no delta
	if delta.Name == nil && delta.Price == nil && delta.Hidden == nil && delta.Suspended == nil {
		return nil
	}

	// we're non-nil. Add the provider id
	delta.ProviderID = a.ProviderID

	return delta
}

func stringValueDelta(a, b *wrappers.StringValue) *wrappers.StringValue {
	if b == nil {
		return nil
	}

	if a == nil {
		return &*b
	}

	if a.Value != b.Value {
		return &*b
	}

	return nil
}

func boolValueDelta(a, b *wrappers.BoolValue) *wrappers.BoolValue {
	if b == nil {
		return nil
	}

	if a == nil {
		return &*b
	}

	if a.Value != b.Value {
		return &*b
	}

	return nil
}

func doubleValueDelta(a, b *wrappers.DoubleValue) *wrappers.DoubleValue {
	if b == nil {
		return nil
	}

	if a == nil {
		return &*b
	}

	if a.Value != b.Value {
		return &*b
	}

	return nil
}
