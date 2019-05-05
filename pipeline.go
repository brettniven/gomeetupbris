package gomeetupbris

import (
	"github.com/sirupsen/logrus"
	"gitlab.com/priceshield/agent-gateway/api"
	"strconv"
	"time"
)

// pipeline models the data gathering pipeline
type pipeline struct {
	in                chan<- pipelineItem
	sink              <-chan pipelineItem
	cache             *cache
	liveEventsFetcher *liveEventsFetcher
	merger            *merger
	publisher         *publisher
	steps             []pipelineStep
}

// pipelineStep models a step in the pipeline
type pipelineStep interface {
	start()
	stop()
}

type pipelineResult struct {
	numEvents  int
	numSuccess int
	numErrs    int
	duration   time.Duration
}

func newPipeline(cache *cache, liveSportsEndpoint string, agentGatewayClient api.AgentGatewayClient, deltasOnly bool) *pipeline {

	// liveEventsFetcher fetches provider specific live events, converts them to the agent model and add these to the provided channel
	liveEventsFetcher := newLiveEventsFetcher(liveSportsEndpoint)

	// the sink chan is where the pipeline finishes. Both success and error results are sent here
	sink := make(chan pipelineItem, 100)

	toConvert := make(chan pipelineItem, 100)
	converted := make(chan pipelineItem, 100)
	// converter converts provider state to the common format
	converter := newConverter(toConvert, converted)

	merged := make(chan pipelineItem, 100)
	// merger merges new state with previous state, and outputs changes to the provided output channel
	merger := newMerger(converted, merged, sink)

	// publisher receives outgoing deltas (from deltaOutputChan), publishes them to the agent-gateway service, and publishes the result to the sink
	publisher := newPublisher(merged, sink, agentGatewayClient, deltasOnly)

	steps := []pipelineStep{
		converter,
		merger,
		publisher,
	}

	return &pipeline{
		in:                toConvert,
		sink:              sink,
		cache:             cache,
		liveEventsFetcher: liveEventsFetcher,
		merger:            merger,
		publisher:         publisher,
		steps:             steps,
	}
}

type pipelineItem struct {
	providerEvent ProviderEvent   // a provider event
	prevState     *api.AgentEvent // The state from any previous run
	converted     api.AgentEvent  // the converted event
	merged        *api.AgentEvent // the merged event
	delta         *api.AgentEvent // the delta from the previous state
	err           error           // any error that occurs in the pipeline
}

func (p *pipeline) run() pipelineResult {

	logrus.Info("running_pipeline")

	startTime := time.Now()

	// start all pipeline steps. chans at the ready
	for _, ps := range p.steps {
		go ps.start()
	}
	// ensure we stop them all when we finish
	defer func() {
		for _, ps := range p.steps {
			go ps.stop()
		}
	}()

	pr := pipelineResult{}

	// fetch live events. This has to happen first, in order to know how many pipeline items to later wait for
	providerEvents, err := p.liveEventsFetcher.run()
	if err != nil {
		pr.numErrs++
		logrus.WithError(err).Error("fetch_live_events")
		return pr
	}

	for _, pe := range providerEvents {

		// find the related cache item (if any)
		// do this here whilst in a single goroutine before we fanout
		var prevState *api.AgentEvent
		provEventID := strconv.FormatInt(pe.MasterEventID, 10)
		ae, ok := p.cache.get(provEventID)
		if ok {
			prevState = &ae
		}

		// each element can enter the pipeline asynchronously
		go func(pe ProviderEvent) {
			pi := pipelineItem{
				providerEvent: pe,
				prevState:     prevState,
			}
			// pass to the first step in the pipeline
			p.in <- pi
		}(pe)
	}

	pr.numEvents = len(providerEvents)

	// wait for the pipeline items to finish. note that both successes and failures are written to the sink chan
	// read the sink items in a single goroutine. This allows safe writing to the cache
	rcvd := 0
	for pi := range p.sink {
		rcvd++
		if pi.err != nil {
			logrus.WithError(pi.err).Error()
			pr.numErrs++
			if rcvd == pr.numEvents {
				break
			}
			continue
		}

		pr.numSuccess++

		p.cache.set(*pi.merged)

		if rcvd == pr.numEvents {
			break
		}
	}

	pr.duration = time.Since(startTime)

	logrus.WithField("numEvents", pr.numEvents).
		WithField("numSuccess", pr.numSuccess).
		WithField("numErrs", pr.numErrs).
		WithField("duration (ms)", pr.duration.Nanoseconds()/int64(time.Millisecond)).
		Info("pipeline_completed")

	return pr
}
