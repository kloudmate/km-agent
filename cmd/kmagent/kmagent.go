//go:build ignore

package kmagent

func main() {
	println("This file is ignored unless built with tag 'ignore'")
}

//package main
//
//import (
//	"context"
//	"fmt"
//	"github.com/kloudmate/km-agent/internal/config"
//	"github.com/kloudmate/km-agent/internal/updater"
//	"go.opentelemetry.io/collector/otelcol"
//	"go.opentelemetry.io/collector/service"
//	"log"
//	"log/slog"
//	"os"
//
//	bgsvc "github.com/kardianos/service"
//	cli "github.com/urfave/cli/v2"
//
//	"github.com/kloudmate/km-agent/internal/agent"
//)
//
//var logger bgsvc.Logger
//
//type agentApp struct {
//	ctx        context.Context
//	cancelFunc context.CancelFunc
//	logger     *slog.Logger
//	cfg        *config.Config  // Agent-specific config
//	svcConfig  *service.Config // Service configuration
//
//	collector       *otelcol.Collector // The running collector instance
//	remoteUpdater   *updater.ConfigUpdater
//	shutdownSignals chan os.Signal
//	restartChan     chan struct{} // Receives signal from remote updater
//	exitCode        int           // Exit code to use when self-terminating for restart
//}
//
//func main() {
//	var svcConfig = &bgsvc.Config{
//		Name:        "kmagent",
//		DisplayName: "KloudMate Agent",
//		Description: "KloudMate Agent for auto instrumentation",
//		Arguments:   []string{"start"},
//	}
//	prg := &agent.KmAgentService{
//		AppConfig: agent.AppConfig{},
//	}
//	//
//	s, err := bgsvc.New(prg, svcConfig)
//	if err != nil {
//		log.Fatal(err)
//	}
//	errs := make(chan error, 56)
//	logger, err = s.Logger(errs)
//	if err != nil {
//		log.Fatal(err)
//	}
//	prg.Svclogger = logger
//
//	app := &cli.App{
//		Name:     "kmagent",
//		Usage:    "KloudMate Agent for auto instrumentation",
//		Flags:    prg.CliArgs(),
//		Commands: prg.CliCommands(s),
//		Before: func(c *cli.Context) error {
//			fmt.Println("🔍 Preprocessing CLI arguments...")
//			_, err := config.LoadConfig(c)
//			if err != nil {
//				log.Fatal(err)
//			}
//			os.Setenv("KM_API_KEY", prg.AgentCfg.Key)
//			os.Setenv("KM_COLLECTOR_ENDPOINT", prg.AgentCfg.Endpoint)
//
//			_, err = agent.NewKmAgentService()
//			if err != nil {
//				log.Fatal(err)
//			}
//
//			fmt.Println("Starting Program...")
//			err = prg.Start(s)
//			if err != nil {
//				log.Fatal("Failed to start KM Agent ", err)
//			}
//
//			return nil
//		},
//		Action: func(c *cli.Context) error {
//			return cli.ShowAppHelp(c)
//		},
//	}
//	// prg.ApplyAgentConfig(cli.NewContext(app, nil, nil))
//
//	if err := app.Run(os.Args); err != nil {
//		log.Fatal("Failed to run KM Agent", err)
//	}
//}
//
////
////func run(c *cli.Context) error {
////	fmt.Println("Starting Custom OpenTelemetry Collector Agent...")
////
////	// Load configuration with priority: CLI flags > Environment Variables > Config File
////	cfg, err := config.Load(c)
////	if err != nil {
////		return fmt.Errorf("failed to load configuration: %w", err)
////	}
////
////	// Create and start the agent
////	a, err := agent.New(cfg)
////	if err != nil {
////		return fmt.Errorf("failed to create agent: %w", err)
////	}
////
////	// Setup signal handling for graceful shutdown
////	ctx, cancel := context.WithCancel(context.Background())
////	defer cancel()
////
////	signalCh := make(chan os.Signal, 1)
////	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
////
////	go func() {
////		<-signalCh
////		fmt.Println("Received shutdown signal")
////		cancel()
////	}()
////
////	// Start the agent
////	if err := a.Start(ctx); err != nil {
////		return fmt.Errorf("failed to start agent: %w", err)
////	}
////
////	// Wait for shutdown signal
////	<-ctx.Done()
////
////	// Graceful shutdown
////	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
////	defer shutdownCancel()
////
////	if err := a.Shutdown(shutdownCtx); err != nil {
////		return fmt.Errorf("error during shutdown: %w", err)
////	}
////
////	fmt.Println("Agent stopped successfully")
////	return nil
////}
