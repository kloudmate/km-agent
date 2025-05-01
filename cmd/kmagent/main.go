package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/kardianos/service"
	"github.com/kloudmate/km-agent/internal/agent"
	"github.com/kloudmate/km-agent/internal/config"
)

// CLIConfig holds all CLI-related configurations
type CLIConfig struct {
	ConfigPath          string
	ExporterEndpoint    string
	APIKey              string
	ConfigCheckInterval int
	DockerMode          bool
}

// Program represents the main application state and configuration
type Program struct {
	// Configuration
	cfg *config.Config

	// CLI Configuration
	cliConfig CLIConfig

	// Core components
	kmAgent *agent.Agent
	logger  *zap.SugaredLogger

	// Application lifecycle
	ctx        context.Context
	cancelFunc context.CancelFunc
	sigChan    chan os.Signal
	quitCh     chan struct{}
	errCh      chan error
	wg         *sync.WaitGroup

	// CLI app
	app *cli.App
}

func (p *Program) run() {
	// Initialize the agent
	defer p.wg.Done()
	p.logger.Info("Service is running, Docker mode: ", p.cfg.Agent.DockerMode)
	for err := range p.errCh {
		if err != nil {
			p.logger.Errorf("Error: %v", err)
			// stop
			continue
		}
		err = p.kmAgent.StartAgent(p.ctx)
		if err != nil {
			p.logger.Error("Error starting agent: %v", err)
		}

	}
	p.logger.Info("In Run loop")
	//
	//if err := p.kmAgent.StartAgent(p.ctx); err != nil {
	//	return fmt.Errorf("failed to start agent: %v", err)
	//}
	//
	//// Wait for shutdown signal
	//sig := <-p.sigChan
	//p.logger.Infof("Received signal: %v", sig)
	//
	//// Shutdown
	//if err := p.kmAgent.Shutdown(p.ctx); err != nil {
	//	p.logger.Errorf("Error during shutdown: %v", err)
	//	return err
	//}
	//
	//p.logger.Info("Agent successfully shut down")
	//return nil

}

func (p *Program) Start(s service.Service) error {
	p.logger.Info("Starting service")
	p.wg.Add(1)
	go p.run()
	p.errCh <- nil
	p.logger.Info("Service started")
	return nil
}

func (p *Program) Stop(s service.Service) error {
	p.logger.Info("Stopping service")
	close(p.quitCh)
	close(p.sigChan)
	close(p.errCh)

	p.wg.Wait()
	p.logger.Info("Service stopped")
	return nil
}

func (p *Program) Initialize(c *cli.Context) error {
	var err error

	// Load configuration
	p.cfg, err = config.LoadConfig(c)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create agent
	p.kmAgent, err = agent.New(p.cfg, p.logger)
	if err != nil {
		return fmt.Errorf("failed to create agent: %v", err)
	}

	// Setup context and signal handling
	p.ctx, p.cancelFunc = context.WithCancel(context.Background())
	signal.Notify(p.sigChan, syscall.SIGINT, syscall.SIGTERM)

	return nil
}

// Run starts the program and blocks until shutdown
//func (p *Program) Run() {
//	defer p.wg.Done()
//	p.logger.Info("Service is running")
//	// Start the agent
//	if err := p.kmAgent.StartAgent(p.ctx); err != nil {
//		p.errCh <- fmt.Errorf("failed to start agent: %v", err)
//		return
//	}
//
//	// Wait for shutdown signal
//	sig := <-p.sigChan
//	p.logger.Infof("Received signal: %v", sig)
//}

// Shutdown gracefully shuts down the program
func (p *Program) Shutdown() {
	if p.cancelFunc != nil {
		p.cancelFunc()
	}
	p.wg.Wait()
}

func makeService(p *Program) service.Service {
	svcConfig := &service.Config{
		Name:        "kmagent",
		DisplayName: "KloudMate Agent",
		Description: "KloudMate Agent for OpenTelemetry auto instrumentation",
	}
	svc, err := service.New(p, svcConfig)
	if err != nil {
		p.logger.Error("Error creating service: %v", err)
	}
	return svc
}

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}

	wg := &sync.WaitGroup{}
	// Create program instance
	program := &Program{
		logger:  logger.Sugar(),
		sigChan: make(chan os.Signal, 1),
		quitCh:  make(chan struct{}),
		errCh:   make(chan error),
		wg:      wg,
	}

	// Create CLI app
	program.app = &cli.App{
		Name:  "kmagent",
		Usage: "KloudMate OpenTelemetry collector agent",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Usage:       "Path to the collector configuration file",
				EnvVars:     []string{"KM_AGENT_CONFIG"},
				Destination: &program.cliConfig.ConfigPath,
			},
			&cli.StringFlag{
				Name:        "exporter-endpoint",
				Usage:       "OpenTelemetry exporter endpoint",
				EnvVars:     []string{"KM_COLLECTOR_ENDPOINT"},
				Destination: &program.cliConfig.ExporterEndpoint,
			},
			&cli.StringFlag{
				Name:        "api-key",
				Usage:       "API key for authentication",
				EnvVars:     []string{"KM_API_KEY"},
				Destination: &program.cliConfig.APIKey,
			},
			&cli.IntFlag{
				Name:        "config-check-interval",
				Usage:       "Interval in seconds to check for config updates",
				Value:       300, // 5 minutes default
				EnvVars:     []string{"KM_CONFIG_CHECK_INTERVAL"},
				Destination: &program.cliConfig.ConfigCheckInterval,
			},
			&cli.BoolFlag{
				Name:        "docker-mode",
				Usage:       "Run in Docker mode with specialized configuration",
				EnvVars:     []string{"KM_AGENT_MODE"},
				Destination: &program.cliConfig.DockerMode,
			},
		},
	}

	svc := makeService(program)

	// Setup commands
	program.app.Commands = []*cli.Command{
		{
			Name:  "install",
			Usage: "Install the agent as a system service",
			Action: func(c *cli.Context) error {
				program.logger.Info("Installing agent as a system service...")
				return svc.Install()
			},
		},
		{
			Name:  "uninstall",
			Usage: "Uninstall the agent service",
			Action: func(c *cli.Context) error {
				program.logger.Info("Uninstalling agent service...")
				return svc.Uninstall()
			},
		},
		{
			Name:  "start",
			Usage: "Start the agent service",
			Action: func(c *cli.Context) error {
				program.logger.Info("Starting agent service...")
				return svc.Start()
			},
		},
		{
			Name:  "stop",
			Usage: "Stop the agent service",
			Action: func(c *cli.Context) error {
				program.logger.Info("Stopping agent service...")
				return svc.Stop()
			},
		},
		{
			Name:  "run",
			Usage: "Run the agent as a standalone application",
			Action: func(c *cli.Context) error {
				os.Setenv("KM_COLLECTOR_ENDPOINT", program.cliConfig.ExporterEndpoint)
				os.Setenv("KM_API_KEY", program.cliConfig.APIKey)
				if err := program.Initialize(c); err != nil {
					return err
				}
				err = svc.Run()
				if err != nil {
					program.logger.Fatal(err)
				}
				return nil
			},
		},
	}

	// Default action shows help
	program.app.Action = func(c *cli.Context) error {
		return cli.ShowAppHelp(c)
	}

	// Run the CLI app
	if err := program.app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
