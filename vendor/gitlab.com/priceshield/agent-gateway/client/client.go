package client

import (
	"context"
	"gitlab.com/priceshield/agent-gateway/api"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// NewClient creates a new client and attempts to establish a connection
func NewClient(addr string) (api.AgentGatewayClient, error) {
	return connectToAPIGatewayService(addr)
}

func connectToAPIGatewayService(addr string) (api.AgentGatewayClient, error) {

	logrus.Info("connecting_to_agentgateway_service")
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := api.NewAgentGatewayClient(conn)

	for i := 1; i <= 10; i++ {
		err = connect(client)
		if err == nil {
			logrus.WithField("attempt", i).Info("connected_to_agentgateway")
			break
		}
		logrus.WithField("attempt", i).Debug("agentgateway_connect_failure")
		time.Sleep(time.Second * 2)
	}

	return client, err
}

func connect(client api.AgentGatewayClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()
	_, err := client.Ping(ctx, &api.PingRequest{})
	if err != nil {
		return err
	}
	logrus.Info("connected_to_agentgateway")
	return nil
}
