//go:build k8s
// +build k8s

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kloudmate/km-agent/internal/k8sagent"
	"go.uber.org/zap"
)

func main() {
	// Set up the main context for the application, which can be cancelled for shutdown.
	appCtx, cancelAppCtx := context.WithCancel(context.Background())
	defer cancelAppCtx()

	agent, err := k8sagent.NewK8sAgent()
	if err != nil {
		log.Fatal(err)
	}

	// Handle OS signals for graceful shutdown.
	handleSignals(cancelAppCtx, agent.Logger)

	agent.FilterValidResources(appCtx, agent.Logger)
	// agent.Logger.Infof("cluster in config : %s\n", agent.Cfg.Monitoring.ClusterName)

	otelCfg, err := k8sagent.GenerateCollectorConfig(agent.Cfg) // Generate otel config based on our agent-config

	if err != nil {
		log.Fatalf("agent could not generate collector config : %s", err.Error())
	}

	if err = agent.StartOTelWithGeneratedConfig(otelCfg); err != nil {
		log.Fatalf("agent could not be started with current config : %s", err.Error())
	}

	defer func() {
		// Ensure logger is synced before exit to flush any buffered logs.
		if syncErr := agent.Logger.Sync(); syncErr != nil && syncErr.Error() != "sync /dev/stdout: invalid argument" {
			// Ignore "invalid argument" error for stdout/stderr
			agent.Logger.Errorf("Failed to sync logger: %v\n", syncErr)
		}
	}()

	for sig := range agent.Errs {
		agent.Logger.Infof("status : %v \n Gracefully shutting down", sig)

		// Deregister the km-agent from kloudmate api if required / turn of health checks any other cleanup task
		os.Exit(0)
	}
}

// handleSignals sets up a signal handler to gracefully shut down the agent.
func handleSignals(cancelFunc context.CancelFunc, l *zap.SugaredLogger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		l.Warnf("Received signal %s, initiating shutdown...", sig)
		cancelFunc() // Cancel the main context to trigger graceful shutdown
	}()
}
