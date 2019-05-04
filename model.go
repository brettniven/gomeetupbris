package gomeetupbris

import "time"

// LiveEndpointResponse models the provider response
type LiveEndpointResponse struct {
	EventTypes []EventType `json:"EventTypes"`
}

// EventType models a sport
type EventType struct {
	EventTypeDesc string          `json:"EventTypeDesc"` // e.g. Basketball
	EventTypeSlug string          `json:"EventTypeSlug"` // e.g. basketball
	Events        []ProviderEvent `json:"Events"`
}

// ProviderEvent models the provider event
type ProviderEvent struct {
	MasterEventID             int64          `json:"MasterEventID"`
	MasterEventName           string         `json:"MasterEventName"`           // e.g. Melbourne United v Cairns Taipans
	DisplayMasterCategoryName string         `json:"DisplayMasterCategoryName"` // e.g NBL
	FeaturedMarket            FeaturedMarket `json:"FeaturedMarket"`
	Sport                     string         // not json, populated when iterating
}

// FeaturedMarket models the main market
type FeaturedMarket struct {
	EventID             int64     `json:"EventID"`
	EventName           string    `json:"EventName"`           // e.g. Head to Head
	AdvertisedStartTime time.Time `json:"AdvertisedStartTime"` // e.g 2019-02-14T08:50:00Z
	IsOpenForBetting    bool      `json:"IsOpenForBetting"`
	Outcomes            []Outcome `json:"Outcomes"`
}

// Outcome models a selection
type Outcome struct {
	FixedMarketID int64   `json:"FixedMarketID"` // this seems to be unique per selection
	OutcomeID     int64   `json:"OutcomeID"`     // this is unique within the market. i.e. 1/2/3 for teamA, draw, teamB
	OutcomeName   string  `json:"OutcomeName"`   // selection name
	Price         float64 `json:"Price"`         // decimal price
}
