package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/kloudmate/km-agent/internal/collector"
	"github.com/kloudmate/km-agent/internal/config"
	"github.com/kloudmate/km-agent/internal/updater"
)

// Agent represents the OpenTelemetry agent
type Agent struct {
	cfg            *config.Config
	logger         *zap.SugaredLogger
	collector      *collector.Collector
	updater        *updater.ConfigUpdater
	shutdownSignal chan struct{}
	wg             sync.WaitGroup
	mu             sync.Mutex
	isRunning      bool
}

// New creates a new Agent instance
func New(cfg *config.Config, logger *zap.SugaredLogger) (*Agent, error) {
	// Create collector
	otelCollector, err := collector.NewCollector(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create collector: %w", err)
	}

	// Create config updater
	configUpdater := updater.NewConfigUpdater(cfg, logger)

	return &Agent{
		cfg:            cfg,
		logger:         logger,
		collector:      otelCollector,
		updater:        configUpdater,
		shutdownSignal: make(chan struct{}),
	}, nil
}

// StartAgent starts the agent
func (a *Agent) StartAgent(ctx context.Context) error {
	a.mu.Lock()
	if a.isRunning {
		a.mu.Unlock()
		return fmt.Errorf("agent is already running")
	}
	a.isRunning = true
	a.mu.Unlock()

	// Start the collector
	if err := a.collector.Start(ctx); err != nil {
		return fmt.Errorf("failed to start collector: %w", err)
	}

	// Start config update checker in a separate goroutine
	a.wg.Add(1)
	go a.runConfigUpdateChecker(ctx)

	a.logger.Info("Agent started successfully")
	return nil
}

// Shutdown gracefully shuts down the agent
func (a *Agent) Shutdown(ctx context.Context) error {
	a.mu.Lock()
	if !a.isRunning {
		a.mu.Unlock()
		return nil
	}
	a.isRunning = false
	a.mu.Unlock()

	// Signal the update checker to stop
	close(a.shutdownSignal)

	// Wait for goroutines to finish
	waitCh := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(waitCh)
	}()

	select {
	case <-waitCh:
		// All goroutines finished
	case <-ctx.Done():
		return ctx.Err()
	}

	// Shutdown the collector
	if err := a.collector.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown collector: %w", err)
	}

	a.logger.Info("Agent shut down successfully")
	return nil
}

// runConfigUpdateChecker periodically checks for configuration updates
func (a *Agent) runConfigUpdateChecker(ctx context.Context) {
	defer a.wg.Done()

	// Skip if no config update URL is configured
	if a.cfg.Agent.ConfigUpdateURL == "" {
		a.logger.Info("Config update URL not configured, skipping config update checks")
		return
	}

	ticker := time.NewTicker(time.Duration(a.cfg.Agent.ConfigCheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.logger.Debug("Checking for configuration updates")
			restartRequired, newConfig, err := a.updater.CheckForUpdates(ctx)
			if err != nil {
				a.logger.Errorf("Failed to check for config updates: %v", err)
				continue
			}

			if newConfig != nil {
				a.logger.Info("New configuration received")

				// Update the collector configuration
				if err := a.collector.UpdateConfig(ctx, newConfig); err != nil {
					a.logger.Errorf("Failed to update collector configuration: %v", err)
					continue
				}

				// If restart is required, restart the collector
				if restartRequired {
					a.logger.Info("Restart required, restarting collector")

					// Shutdown current collector
					if err := a.collector.Shutdown(ctx); err != nil {
						a.logger.Errorf("Failed to shutdown collector for restart: %v", err)
						continue
					}

					// Start collector with new config
					if err := a.collector.Start(ctx); err != nil {
						a.logger.Errorf("Failed to restart collector: %v", err)
						continue
					}

					a.logger.Info("Collector restarted successfully")
				}
			}

		case <-a.shutdownSignal:
			a.logger.Debug("Config update checker stopping")
			return
		case <-ctx.Done():
			a.logger.Debug("Config update checker context canceled")
			return
		}
	}
}
