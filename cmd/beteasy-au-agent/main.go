package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brettniven/gomeetupbris"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	agentgatewayclient "gitlab.com/priceshield/agent-gateway/client"
)

func main() {

	var c Config
	err := envconfig.Process("", &c)
	if err != nil {
		logrus.WithError(err).Fatal("invalid_config")
	}

	configureLogging(c)

	// dependencies
	agentgatewayClient, err := agentgatewayclient.NewClient(c.AgentGatewayAddr)
	if err != nil {
		logrus.WithError(err).Fatal("agentgateway_connect_failure")
	}

	pollInterval := time.Second * time.Duration(c.PollIntervalSecs)

	agent := gomeetupbris.NewService(agentgatewayClient, c.LiveSportsEndpoint, pollInterval, c.DeltasOnly)

	mux := http.NewServeMux()
	mux.Handle("/", gomeetupbris.NewHandler())

	errs := make(chan error, 1)

	httpAddr := fmt.Sprintf(":%s", c.HTTPPort)
	// serve http
	go func() {
		logrus.WithField("address", httpAddr).Info("serving_http")
		errs <- http.ListenAndServe(httpAddr, mux)
		logrus.Info("finished_serving_http")
	}()

	// start the agent
	go agent.Start()

	// listen for signals to die
	signals := make(chan os.Signal, 1)
	signal.Notify(signals,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	select {
	case err := <-errs:
		logrus.WithError(err).Error("service_error")
	case sig := <-signals:
		logrus.WithField("signal", sig).Warn("shutdown_signal")
	}

	agent.Stop()

	logrus.Info("terminated")
}

func configureLogging(c Config) {
	level, errLevel := logrus.ParseLevel(c.LogLevel)
	if errLevel == nil {
		logrus.SetLevel(level)
	}
	if c.LogFormat == "text" {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	logrus.SetOutput(os.Stdout)
}
