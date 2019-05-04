package gomeetupbris

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// liveEventsFetcher fetches live events from the provider endpoint and parses into structs
type liveEventsFetcher struct {
	liveSportsEndpoint string
	client             *http.Client
}

func newLiveEventsFetcher(liveSportsEndpoint string) *liveEventsFetcher {
	client := newHTTPClient()
	return &liveEventsFetcher{
		liveSportsEndpoint: liveSportsEndpoint,
		client:             client,
	}
}

func newHTTPClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 1,
		},
		Timeout: time.Duration(10) * time.Second,
	}

	return client
}

func (p *liveEventsFetcher) run() ([]ProviderEvent, error) {
	logrus.Info("getting_live_events")
	resp, err := p.getLiveEvents()
	if err != nil {
		return nil, errors.Wrap(err, "get_live_events")
	}

	pes := make([]ProviderEvent, 0)
	count := 0
	for _, et := range resp.EventTypes {
		sport := et.EventTypeDesc
		for _, pe := range et.Events {
			count++
			pe.Sport = sport
			pes = append(pes, pe)
		}
	}
	logrus.WithField("count", count).Info("finished_getting_live_events")
	return pes, nil
}

func (p *liveEventsFetcher) getLiveEvents() (*LiveEndpointResponse, error) {

	endPoint := p.liveSportsEndpoint

	req, err := http.NewRequest("GET", endPoint, nil)
	if err != nil {
		return nil, errors.Wrap(err, "request_construct_failed")
	}

	response, err := p.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "get_failed")
	}

	defer func() {
		err := response.Body.Close()
		if err != nil {
			logrus.WithError(err).Error()
		}
	}()

	var rslt = &LiveEndpointResponse{}
	err = json.NewDecoder(response.Body).Decode(rslt)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal_failed")
	}

	return rslt, nil
}

func standardizeMarketName(mn string) string {
	switch strings.TrimSpace(strings.ToLower(mn)) {
	case "head to head":
		fallthrough
	case "h2h":
		fallthrough
	case "match betting":
		return "2-way"
	case "win-draw-win":
		fallthrough
	case "match result":
		return "3-way"
	default:
		return mn
	}
}
