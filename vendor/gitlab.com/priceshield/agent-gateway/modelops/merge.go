package modelops

import (
	"errors"
	"gitlab.com/priceshield/agent-gateway/api"
)

// Merge merges an AgentEvent
func Merge(original, delta *api.AgentEvent) (*api.AgentEvent, error) {

	if original == nil && delta == nil {
		return nil, nil
	}

	if delta == nil {
		cloneOriginal := *original
		return &cloneOriginal, nil
	}

	if original == nil {
		cloneDelta := *delta // TODO clone properly
		return &cloneDelta, nil
	}

	// if we are here, both are non-nil
	if original.ProviderID != delta.ProviderID {
		return nil, errors.New("events_with_differing_ids")
	}

	cloneOriginal := Clone(original)

	if delta.Name != nil {
		cloneOriginal.Name = delta.Name
	}

	if delta.Sport != nil {
		cloneOriginal.Sport = delta.Sport
	}

	if delta.Competition != nil {
		cloneOriginal.Competition = delta.Competition
	}

	if delta.Hidden != nil {
		cloneOriginal.Hidden = delta.Hidden
	}

	if delta.Suspended != nil {
		cloneOriginal.Suspended = delta.Suspended
	}

	mergeMarkets(cloneOriginal, delta)

	return cloneOriginal, nil
}

func mergeMarkets(cloneOriginal, delta *api.AgentEvent) {

	if delta.Markets == nil {
		return
	}

	if cloneOriginal.Markets == nil {
		cloneOriginal.Markets = delta.Markets
		return
	}

	for _, dm := range delta.Markets {

		om := findMatchingMarket(cloneOriginal.Markets, dm)
		if om == nil {
			// no match. append
			cloneDM := &*dm
			cloneOriginal.Markets = append(cloneOriginal.Markets, cloneDM)
		} else {
			// found a match. merge	to original
			mergeMarket(om, dm)
		}
	}
}

func findMatchingMarket(markets []*api.AgentMarket, market *api.AgentMarket) *api.AgentMarket {
	for _, m := range markets {
		if m.ProviderID == market.ProviderID {
			return m
		}
	}
	return nil
}

func mergeMarket(cloneOriginal, delta *api.AgentMarket) {

	if delta.Name != nil {
		cloneOriginal.Name = delta.Name
	}

	if delta.Hidden != nil {
		cloneOriginal.Hidden = delta.Hidden
	}

	if delta.Suspended != nil {
		cloneOriginal.Suspended = delta.Suspended
	}

	mergeOptions(cloneOriginal, delta)
}

func mergeOptions(cloneOriginal, delta *api.AgentMarket) {

	if delta.Options == nil {
		return
	}

	if cloneOriginal.Options == nil {
		cloneOriginal.Options = delta.Options
		return
	}

	for _, do := range delta.Options {

		oo := findMatchingOption(cloneOriginal.Options, do)
		if oo == nil {
			// no match. append
			cloneDO := &*do
			cloneOriginal.Options = append(cloneOriginal.Options, cloneDO)
		} else {
			// found a match. merge	to original
			mergeOption(oo, do)
		}
	}
}

func findMatchingOption(options []*api.AgentOption, option *api.AgentOption) *api.AgentOption {
	for _, o := range options {
		if o.ProviderID == option.ProviderID {
			return o
		}
	}
	return nil
}

func mergeOption(cloneOriginal, delta *api.AgentOption) {

	if delta.Name != nil {
		cloneOriginal.Name = delta.Name
	}

	if delta.Price != nil {
		cloneOriginal.Price = delta.Price
	}

	if delta.Hidden != nil {
		cloneOriginal.Hidden = delta.Hidden
	}

	if delta.Suspended != nil {
		cloneOriginal.Suspended = delta.Suspended
	}
}
