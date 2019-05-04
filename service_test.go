package gomeetupbris

import (
	"context"
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"gitlab.com/priceshield/agent-gateway/api"
	"google.golang.org/grpc"
)

type serviceTest struct {
	Name                     string
	initialState             []api.AgentEvent
	providerEndpointResponse []byte
	expected                 []api.AgentEvent
}

func TestService(t *testing.T) {

	logrus.SetLevel(logrus.InfoLevel)

	scenarios := make([]serviceTest, 0)

	testDir := "./testdata"
	dirs, err := ioutil.ReadDir(testDir)
	if err != nil {
		t.Error(err)
	}

	// each dir under /testdata, is a test
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}

		scenarioDir := testDir + string(filepath.Separator) + d.Name()

		initialState := readAndUnmarshalAEs(t, scenarioDir, "initial_state.json")
		providerEndpointResponse := readPER(t, scenarioDir, "provider_endpoint_resp.json")
		expected := readAndUnmarshalAEs(t, scenarioDir, "expected.json")

		scenario := serviceTest{
			Name:                     d.Name(),
			initialState:             initialState,
			providerEndpointResponse: providerEndpointResponse,
			expected:                 expected,
		}

		scenarios = append(scenarios, scenario)
	}

	// Run the individual test cases
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			runScenario(t, scenario)
		})
	}
}

func runScenario(t *testing.T, scenario serviceTest) {
	// start real http server, mocking the response declared in the test file
	httpSrv := httpServerWithMockResp(t, scenario.providerEndpointResponse)
	defer httpSrv.Close()

	agentGatewayClient := &MockAgentGatewayClient{
		rcvd: make(chan *api.SubmitAgentDeltaRequest, 100),
	}

	// init the service
	liveSportsEndpoint := httpSrv.URL
	svc := NewService(agentGatewayClient, liveSportsEndpoint, time.Hour, true)
	// set initial state as dictated by the test
	svc.setInitialState(scenario.initialState)
	go svc.Start()
	defer svc.Stop()

	actual := make([]api.AgentEvent, 0)

waitRcv:
	for {
		select {
		case r := <-agentGatewayClient.rcvd:
			assert.Equal(t, "beteasy_au", r.AgentID)
			for _, ae := range r.Events {
				actual = append(actual, *ae)
			}
			if len(actual) >= len(scenario.expected) {
				break waitRcv
			}
		case <-time.After(time.Second * 1):
			if len(scenario.expected) > 0 {
				t.Error("timeout waiting for expected results")
			}
			break waitRcv
		}
	}

	// sort so that our comparison is determinate
	sort.Slice(scenario.expected, func(i, j int) bool {
		return scenario.expected[i].ProviderID < scenario.expected[j].ProviderID
	})
	sort.Slice(actual, func(i, j int) bool {
		return actual[i].ProviderID < actual[j].ProviderID
	})

	if !cmp.Equal(scenario.expected, actual) {
		t.Error(cmp.Diff(scenario.expected, actual))
	}
}

func httpServerWithMockResp(t *testing.T, body []byte) *httptest.Server {
	httpSrv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write(body)
			if err != nil {
				t.Error(err)
				return
			}
		}),
	)

	return httpSrv
}

func readAndUnmarshalAEs(t *testing.T, scenarioDir string, fileName string) []api.AgentEvent {
	b, err := ioutil.ReadFile(scenarioDir + string(filepath.Separator) + fileName)
	if err != nil {
		t.Error(err)
		return nil
	}

	if len(b) == 0 {
		// an empty file, meaning the test wants nil
		return nil
	}

	dest := make([]api.AgentEvent, 0)
	err = json.Unmarshal(b, &dest)
	if err != nil {
		t.Error(err)
		return nil
	}

	return dest
}

func readPER(t *testing.T, scenarioDir string, fileName string) []byte {
	b, err := ioutil.ReadFile(scenarioDir + string(filepath.Separator) + fileName)
	if err != nil {
		t.Error(err)
		return nil
	}

	return b
}

type MockAgentGatewayClient struct {
	rcvd chan *api.SubmitAgentDeltaRequest
}

func (magc *MockAgentGatewayClient) Ping(ctx context.Context, in *api.PingRequest, opts ...grpc.CallOption) (*api.PingResponse, error) {
	return &api.PingResponse{}, nil
}

// SubmitAgentDelta submits an agent delta
func (magc *MockAgentGatewayClient) SubmitAgentDelta(ctx context.Context, in *api.SubmitAgentDeltaRequest, opts ...grpc.CallOption) (*api.SubmitAgentDeltaResponse, error) {
	magc.rcvd <- in
	return &api.SubmitAgentDeltaResponse{}, nil
}
