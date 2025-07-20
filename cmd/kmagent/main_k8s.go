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
	handleSignals(cancelAppCtx, agent)

	agent.FilterValidResources(appCtx, agent.Logger)
	// agent.Logger.Infof("cluster in config : %s\n", agent.Cfg.Monitoring.ClusterName)

	otelCfg, err := k8sagent.GenerateCollectorConfig(agent.Cfg)

	if err != nil {
		log.Fatalf("agent could not generate collector config : %s", err.Error())
	}

	if err = agent.StartAgent(appCtx, otelCfg); err != nil {
		log.Fatalf("agent could not be started with current config : %s", err.Error())
	}

	defer func() {
		// Ensure logger is synced before exit to flush any buffered logs.
		if syncErr := agent.Logger.Sync(); syncErr != nil && syncErr.Error() != "sync /dev/stdout: invalid argument" {
			// Ignore "invalid argument" error for stdout/stderr
			agent.Logger.Errorf("Failed to sync logger: %v\n", syncErr)
		}
	}()
}

// handleSignals sets up a signal handler to gracefully shut down the agent.
func handleSignals(cancelFunc context.CancelFunc, agent *k8sagent.K8sAgent) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		agent.Logger.Warnf("Received signal %s, initiating shutdown...", sig)
		cancelFunc()
		agent.Stop()
	}()
}
