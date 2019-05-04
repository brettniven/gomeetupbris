package main

// Config models the service config
type Config struct {
	LogLevel           string `envconfig:"LOG_LEVEL" default:"Info"`
	LogFormat          string `envconfig:"LOG_FORMAT"`
	HTTPPort           string `envconfig:"HTTP_PORT" default:"8080"`
	PollIntervalSecs   int    `envconfig:"POLL_INTERVAL_SECS" default:"20"`
	LiveSportsEndpoint string `envconfig:"LIVE_SPORTS_ENDPOINT" default:"https://api.beteasy.com.au/WebEventAPI/sports/live/"`
	AgentGatewayAddr   string `envconfig:"AGENT_GATEWAY_ADDR" default:"agent-gateway:50051"`
	DeltasOnly         bool   `envconfig:"DELTAS_ONLY" default:"true"` // when horizontally scaling, this should be set to false
}
