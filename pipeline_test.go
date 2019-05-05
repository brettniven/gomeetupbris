package gomeetupbris

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gitlab.com/priceshield/agent-gateway/api"
	"testing"
)

type pipelineTest struct {
	name                     string
	providerEndpointResponse []byte
	mergeErr                 error
	deltaErr                 error
	agentGatewayErr          error
	expectedEvents           int
	expectedErrs             int
}

func TestPipelineErrConditions(t *testing.T) {

	logrus.SetLevel(logrus.InfoLevel)

	testTable := []pipelineTest{
		{
			name:                     "testMergeError",
			providerEndpointResponse: readPER(t, "./testdata/pipeline/single/", "provider_endpoint_resp.json"),
			mergeErr:                 errors.New("some merge error"),
			expectedEvents:           1,
			expectedErrs:             1,
		},
		{
			name:                     "testDeltaError",
			providerEndpointResponse: readPER(t, "./testdata/pipeline/single/", "provider_endpoint_resp.json"),
			deltaErr:                 errors.New("some delta error"),
			expectedEvents:           1,
			expectedErrs:             1,
		},
		{
			name:                     "testAgentGatewayError",
			providerEndpointResponse: readPER(t, "./testdata/pipeline/single/", "provider_endpoint_resp.json"),
			agentGatewayErr:          errors.New("some agentGateway error"),
			expectedEvents:           1,
			expectedErrs:             1,
		},
		{
			name:                     "testBadEndpointJson",
			providerEndpointResponse: readPER(t, "./testdata/pipeline/err_bad_json/", "provider_endpoint_resp.json"),
			agentGatewayErr:          errors.New("some agentGateway error"),
			expectedEvents:           0,
			expectedErrs:             1,
		},
	}

	for _, pt := range testTable {
		t.Run(pt.name, func(t *testing.T) {
			runPipelineTest(t, pt)
		})
	}

}

func runPipelineTest(t *testing.T, pt pipelineTest) {
	// start real http server, mocking the response declared in the test file
	httpSrv := httpServerWithMockResp(t, pt.providerEndpointResponse)
	defer httpSrv.Close()
	liveSportsEndpoint := httpSrv.URL

	// mock merge err, if applicable
	if pt.mergeErr != nil {
		origMergeFunc := Merge
		defer func() {
			Merge = origMergeFunc
		}()
		Merge = func(original, delta *api.AgentEvent) (*api.AgentEvent, error) {
			return nil, pt.mergeErr
		}
	}

	// mock delta err, if applicable
	if pt.deltaErr != nil {
		origDeltaFunc := Delta
		defer func() {
			Delta = origDeltaFunc
		}()
		Delta = func(original, delta *api.AgentEvent) (*api.AgentEvent, error) {
			return nil, pt.deltaErr
		}
	}

	agentGatewayClient := &MockAgentGatewayClient{
		rcvd: make(chan *api.SubmitAgentDeltaRequest, 100),
	}
	// mock agentGatewayErr, if applicable
	if pt.agentGatewayErr != nil {
		agentGatewayClient.mockSubmitAgentDeltaArr = pt.agentGatewayErr
	}

	p := newPipeline(newCache(), liveSportsEndpoint, agentGatewayClient, true)
	pr := p.run()

	assert.Equal(t, pt.expectedEvents, pr.numEvents)
	assert.Equal(t, pt.expectedErrs, pr.numErrs)
}
