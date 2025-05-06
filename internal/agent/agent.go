package agent

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"sync"
	"time"

	"github.com/kloudmate/km-agent/internal/config"
	"github.com/kloudmate/km-agent/internal/updater"
	"go.opentelemetry.io/collector/otelcol"
	"go.uber.org/zap"
)

// Agent represents the OpenTelemetry agent
type Agent struct {
	cfg            *config.Config
	logger         *zap.SugaredLogger
	collector      *otelcol.Collector
	colSettings    otelcol.CollectorSettings
	updater        *updater.ConfigUpdater
	shutdownSignal chan struct{}
	wg             sync.WaitGroup
	mu             sync.Mutex
	isRunning      bool
}

// New creates a new Agent instance
func New(cfg *config.Config, logger *zap.SugaredLogger) (*Agent, error) {
	// Create collector
	otelCollector, err := NewCollector(cfg)
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
		return fmt.Errorf("collector already running")
	}
	a.isRunning = true
	a.mu.Unlock()

	a.wg.Add(2)
	go func() {
		defer a.wg.Done()
		if err := a.StartCollector(ctx); err != nil {
			a.mu.Lock()
			a.isRunning = false
			a.mu.Unlock()
			a.logger.Errorf("Failed to start collector: %v", err)
		}
	}()
	go func() {
		defer a.wg.Done()
		a.runConfigUpdateChecker(ctx)
	}()

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
	a.collector.Shutdown()
	a.logger.Info("Agent shut down successfully")
	return nil
}

func (a *Agent) StartCollector(ctx context.Context) error {
	err := a.collector.Run(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (a *Agent) StopCollector() {
	a.collector.Shutdown()
}

func (a *Agent) UpdateConfig(ctx context.Context, newConfig map[string]interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Convert config to YAML and write to file
	configYAML, err := yaml.Marshal(newConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal new config to YAML: %w", err)
	}

	// Create a temporary file
	tempFile := a.cfg.OtelConfigPath + ".new"
	if err := os.WriteFile(tempFile, configYAML, 0644); err != nil {
		return fmt.Errorf("failed to write new config to temporary file: %w", err)
	}

	// Rename the temporary file to the actual config file (atomic operation)
	if err := os.Rename(tempFile, a.cfg.OtelConfigPath); err != nil {
		return fmt.Errorf("failed to replace config file: %w", err)
	}

	a.logger.Info("Successfully updated collector configuration")
	return nil
}

// runConfigUpdateChecker periodically checks for configuration updates
func (a *Agent) runConfigUpdateChecker(ctx context.Context) {
	// Skip if no config update URL is configured
	if a.cfg.ConfigUpdateURL == "" {
		a.logger.Info("Config update URL not configured, skipping config update checks")
		return
	}

	ticker := time.NewTicker(time.Duration(a.cfg.ConfigCheckInterval) * time.Second)
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
				if err := a.UpdateConfig(ctx, newConfig); err != nil {
					a.logger.Errorf("Failed to update collector configuration: %v", err)
					continue
				}

				// If restart is required, restart the collector
				if restartRequired {
					a.logger.Info("Restart required, restarting collector")

					// Shutdown current collector
					a.StopCollector()

					// Start collector with new config
					if err := a.StartCollector(ctx); err != nil {
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
