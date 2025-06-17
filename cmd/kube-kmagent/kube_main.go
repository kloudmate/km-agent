package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kloudmate/km-agent/internal/kube"
)

func main() {
	fmt.Println("Doing something")
	ctx := context.Background()

	agent, err := kube.NewKubeAgent()
	if err != nil {
		log.Fatal(err)
	}

	agent.FilterValidResources(ctx, agent.Logger)
	agent.Logger.Infof("Loaded config for cluster: %s\n", agent.Cfg.Monitoring.ClusterName)

	// validConfig, err := kube.FilterValidResources(ctx, client, cfg)
	// if err != nil {
	// 	log.Fatal("Validation error:", err)
	// }

	otelCfg, err := kube.GenerateCollectorConfig(agent.Cfg) // Generate otel config based on our agent-config

	if err = agent.StartOTelWithGeneratedConfig(otelCfg); err != nil {
		log.Fatalf("agent could not be started with current config : %s", err.Error())
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		agent.Errs <- fmt.Errorf("%s", <-c)
	}()

	for sig := range agent.Errs {
		agent.Logger.Errorf("status : %v \n Gracefully shutting down", sig)

		// Deregister the km-agent from kloudmate servers / turn of health checks
		os.Exit(0)
	}
}
