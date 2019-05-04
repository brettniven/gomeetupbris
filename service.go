package gomeetupbris

import (
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.com/priceshield/agent-gateway/api"
)

// BetEasyAgentService models the agent
type BetEasyAgentService struct {
	cache              *cache
	liveSportsEndpoint string
	agentGatewayClient api.AgentGatewayClient
	pollDoneCh         chan bool
	pollInterval       time.Duration
	deltasOnly         bool
}

// NewService creates a new servicez
func NewService(agentGatewayClient api.AgentGatewayClient, liveSportsEndpoint string, pollInterval time.Duration, deltasOnly bool) *BetEasyAgentService {

	cache := newCache()

	a := &BetEasyAgentService{
		cache:              cache,
		liveSportsEndpoint: liveSportsEndpoint,
		agentGatewayClient: agentGatewayClient,
		pollInterval:       pollInterval,
		pollDoneCh:         make(chan bool),
		deltasOnly:         deltasOnly,
	}

	return a
}

// Stop stops the agent
func (svc *BetEasyAgentService) Stop() {
	svc.pollDoneCh <- true
}

func (svc *BetEasyAgentService) runOnce() pipelineResult {
	p := newPipeline(svc.cache, svc.liveSportsEndpoint, svc.agentGatewayClient, svc.deltasOnly)
	return p.run()
}

// Start starts the agent
func (svc *BetEasyAgentService) Start() {

	defer logrus.Info("live_events_poller_exited")

	// run once before starting the ticker
	svc.runOnce()

	t := time.NewTicker(svc.pollInterval).C
	for {
		select {
		case <-t:
			svc.runOnce()
		case <-svc.pollDoneCh:
			return
		}
	}

}

func (svc *BetEasyAgentService) setInitialState(aes []api.AgentEvent) {
	for _, ae := range aes {
		svc.cache.set(ae)
	}
}
