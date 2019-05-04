package gomeetupbris

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gitlab.com/priceshield/agent-gateway/api"
	"time"
)

// publisher publishes pipeline items (deltas or full merge, based on config), to the agent gateway
type publisher struct {
	in                 <-chan pipelineItem
	sink               chan<- pipelineItem // this is the last item in the pipeline, so only the sink to send to
	agentGatewayClient api.AgentGatewayClient
	deltasOnly         bool
	doneChan           chan bool
}

// newPublisher creates a new publisher
func newPublisher(in <-chan pipelineItem, sink chan<- pipelineItem, agentGatewayClient api.AgentGatewayClient, deltasOnly bool) *publisher {

	p := &publisher{
		in:                 in,
		sink:               sink,
		agentGatewayClient: agentGatewayClient,
		deltasOnly:         deltasOnly,
		doneChan:           make(chan bool),
	}

	return p
}

func (p *publisher) start() {

	defer logrus.Info("publisher_exited")

	batch := make([]pipelineItem, 0)
	// send whenever we get to length of 100 or 50ms (whichever comes first)
	maxBatchSendInterval := time.NewTicker(time.Millisecond * 50).C

	for {
		send := false
		select {
		case pi := <-p.in:
			batch = append(batch, pi)
			send = len(batch) >= 100
		case <-maxBatchSendInterval:
			send = len(batch) > 0
		case <-p.doneChan:
			return
		}

		if !send {
			continue
		}

		// send the batch
		err := p.sendBatch(batch)
		for _, pi := range batch {
			if err != nil {
				pi.err = err
			}
			// forward to the sink
			p.sink <- pi
		}

		// clear the batch
		batch = make([]pipelineItem, 0)

	}

}

func (p *publisher) sendBatch(pis []pipelineItem) error {

	agentEventDeltas := make([]*api.AgentEvent, 0)
	for _, pi := range pis {
		if p.deltasOnly {
			// send delta only. should only be used when a single instance running. Minimizes network traffic
			if pi.delta != nil {
				agentEventDeltas = append(agentEventDeltas, pi.delta)
			}
		} else if pi.merged != nil {
			// send full state. should be used when horizontally scaling
			agentEventDeltas = append(agentEventDeltas, pi.merged)
		}
	}

	logrus.WithField("size", len(agentEventDeltas)).Info("sending_batch")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()

	req := &api.SubmitAgentDeltaRequest{
		AgentID: "beteasy_au",
		Events:  agentEventDeltas,
	}

	logrus.WithField("req", req).Debug("submitting")

	_, err := p.agentGatewayClient.SubmitAgentDelta(ctx, req)

	if err != nil {
		return errors.Wrap(err, "submit_agent_delta")
	}

	logrus.Debug("submit_agent_delta_success")
	return nil
}

func (p *publisher) stop() {
	p.doneChan <- true
}
