package main

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2/altsrc"
	"os"
	"sync"
	"time"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/kardianos/service"
	"github.com/kloudmate/km-agent/internal/agent"
	"github.com/kloudmate/km-agent/internal/config"
)

var version = "0.1.0" // Version of the application

// Program represents the main application state and configuration
type Program struct {
	// Configuration
	cfg *config.Config
	// Core components
	kmAgent *agent.Agent
	logger  *zap.SugaredLogger

	// Application lifecycle
	ctx        context.Context
	cancelFunc context.CancelFunc
	//sigChan    chan os.Signal
	//quitCh     chan struct{}
	//errCh      chan error
	wg      *sync.WaitGroup
	version string
}

func (p *Program) run() {
	// Initialize the agent
	defer p.wg.Done()
	p.logger.Info("Service is running, Docker mode: ", p.cfg.DockerMode)
	if err := p.kmAgent.StartAgent(p.ctx); err != nil {
		p.logger.Errorf("Error initially starting agent: %v. The agent might not be running.", err)
		p.cancelFunc()
		return
	}
	<-p.ctx.Done()
	p.logger.Info("Program run method exiting due to context cancellation.")
}

func (p *Program) Start(s service.Service) error {
	p.logger.Info("Starting service")
	p.wg.Add(1)
	go p.run()
	p.logger.Info("Service goroutine started")
	return nil
}

func (p *Program) Stop(s service.Service) error {
	p.logger.Info("Stopping service...")
	// 1. Signal all dependent components to stop
	if p.cancelFunc != nil {
		p.logger.Info("Cancelling program context...")
		p.cancelFunc()
	}
	// 2. Shutdown the agent gracefully
	if p.kmAgent != nil {
		p.logger.Info("Shutting down KloudMate agent...")
		// Provide a new context for shutdown, or use a timeout context
		// If p.ctx is used, it's already canceled, which is fine for agent's Shutdown.
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second) // Example timeout
		defer shutdownCancel()
		if err := p.kmAgent.Shutdown(shutdownCtx); err != nil {
			p.logger.Errorf("Error during agent shutdown: %v", err)
		} else {
			p.logger.Info("KloudMate agent shut down successfully.")
		}
	}
	// 3. Wait for the main program goroutine (p.run) to finish
	p.logger.Info("Waiting for program run goroutine to complete...")
	p.wg.Wait()

	p.logger.Info("Service stopped successfully.")
	return nil
}

func (p *Program) Initialize(c *cli.Context) error {
	var err error
	p.logger.Info("Initializing program...")

	p.ctx, p.cancelFunc = context.WithCancel(context.Background())

	// Load configuration
	err = p.cfg.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create agent
	p.kmAgent, err = agent.New(p.cfg, p.logger, agent.WithVersion(p.version))
	if err != nil {
		return fmt.Errorf("failed to create agent: %v", err)
	}

	p.logger.Info("Program initialized successfully.")

	return nil
}

// Shutdown gracefully shuts down the program
func (p *Program) Shutdown() {
	if p.cancelFunc != nil {
		p.cancelFunc()
	}
	p.wg.Wait()
}

func makeService(p *Program) (service.Service, error) {
	svcConfig := &service.Config{
		Name:        "kmagent",
		DisplayName: "KloudMate Agent",
		Description: "KloudMate Agent for OpenTelemetry auto instrumentation",
	}
	svc, err := service.New(p, svcConfig)
	if err != nil {
		// p.logger.Errorf("Error creating service: %v", err) // Log is fine
		return nil, fmt.Errorf("error creating service: %w", err) // Return error
	}
	return svc, nil
}

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	sugar := logger.Sugar()

	wg := &sync.WaitGroup{}
	// Create program instance
	program := &Program{
		logger:  sugar,
		cfg:     &config.Config{},
		wg:      wg,
		version: version,
	}

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:    "agent-config",
			Usage:   "Path to the agent configuration file",
			EnvVars: []string{"KM_AGENT_CONFIG"},
		},
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "config",
			Usage:       "Path to the collector configuration file",
			EnvVars:     []string{"KM_COLLECTOR_CONFIG"},
			Destination: &program.cfg.OtelConfigPath,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "collector-endpoint",
			Usage:       "OpenTelemetry exporter endpoint",
			Value:       "https://otel.kloudmate.com:4318",
			EnvVars:     []string{"KM_COLLECTOR_ENDPOINT"},
			Destination: &program.cfg.ExporterEndpoint,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "api-key",
			Usage:       "API key for authentication",
			EnvVars:     []string{"KM_API_KEY"},
			Destination: &program.cfg.APIKey,
		}),
		altsrc.NewIntFlag(&cli.IntFlag{
			Name:        "config-check-interval",
			Usage:       "Interval in seconds to check for config updates",
			Value:       30,
			EnvVars:     []string{"KM_CONFIG_CHECK_INTERVAL"},
			Destination: &program.cfg.ConfigCheckInterval,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "update-endpoint",
			Usage:       "API key for authentication",
			Value:       "https://api.kloudmate.com/agents/config-check",
			EnvVars:     []string{"KM_UPDATE_ENDPOINT"},
			Destination: &program.cfg.ConfigUpdateURL,
		}),
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "docker-mode",
			Usage:       "Run in Docker mode with specialized configuration",
			EnvVars:     []string{"KM_DOCKER_MODE"},
			Destination: &program.cfg.DockerMode,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "docker-endpoint",
			Usage:       "API key for authentication",
			EnvVars:     []string{"KM_DOCKER_ENDPOINT"},
			Destination: &program.cfg.DockerEndpoint,
		}),
	}

	// Create CLI app
	app := &cli.App{
		Name:   "kmagent",
		Usage:  "KloudMate OpenTelemetry collector agent",
		Flags:  flags,
		Before: altsrc.InitInputSourceWithContext(flags, altsrc.NewYamlSourceFromFlagFunc("agent-config")),
	}

	// Setup commands
	app.Commands = []*cli.Command{
		{
			Name:  "start",
			Usage: "Start the agent",
			Action: func(c *cli.Context) error {
				if err := program.Initialize(c); err != nil {
					program.logger.Errorf("Failed to initialize program: %v", err)
					return err
				}
				svc, err := makeService(program)
				if err != nil {
					program.logger.Fatalf("Failed to create service: %v", err) // Fatal as we can't run
				}
				program.logger.Info("Attempting to run service...")
				if err := svc.Run(); err != nil {
					program.logger.Fatalf("Failed to run service: %v", err) // Fatal on run error
				}
				program.logger.Info("Service run finished.")
				return nil
			},
		},
	}

	// Default action shows help
	app.Action = func(c *cli.Context) error {
		return cli.ShowAppHelp(c)
	}

	// Run the CLI app
	if err := app.Run(os.Args); err != nil {
		sugar.Errorf("Application run failed: %v", err)
		os.Exit(1)
	}
}
