package gomeetupbris

import (
	"github.com/pkg/errors"
	"gitlab.com/priceshield/agent-gateway/modelops"
)

// Merge is an indirection of the Merge func, making it simple to mock them in tests
var Merge = modelops.Merge

// Delta is an indirection of the Merge func, making it simple to mock them in tests
var Delta = modelops.Delta

// merger merges new state into existing state and also calculates the delta
type merger struct {
	in       <-chan pipelineItem
	out      chan<- pipelineItem
	sink     chan<- pipelineItem // where things go to die
	doneChan chan bool
}

func newMerger(in <-chan pipelineItem, out chan<- pipelineItem, sink chan<- pipelineItem) *merger {
	return &merger{
		in:       in,
		out:      out,
		sink:     sink,
		doneChan: make(chan bool),
	}
}

func (m *merger) start() {
	for {
		select {
		case e := <-m.in:
			go func(pi pipelineItem) {
				err := m.handleMergeAndDelta(&pi)
				if err != nil {
					// add to sink
					pi.err = err
					m.sink <- pi
					return
				}
				// all good. forward on to next step
				m.out <- pi
			}(e)
		case <-m.doneChan:
			return
		}
	}
}

func (m *merger) handleMergeAndDelta(pi *pipelineItem) error {

	merged, err := Merge(pi.prevState, &pi.converted)
	if err != nil {
		return errors.Wrap(err, "merge_error")
	}
	pi.merged = merged

	delta, err := Delta(pi.prevState, merged)
	if err != nil {
		return errors.Wrap(err, "delta_error")
	}
	pi.delta = delta

	return nil
}

func (m *merger) stop() {
	m.doneChan <- true
}
