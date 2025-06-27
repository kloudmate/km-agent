package agent

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"

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
	updater        *updater.ConfigUpdater
	shutdownSignal chan struct{}
	wg             sync.WaitGroup
	collectorMu    sync.Mutex
	isRunning      atomic.Bool
}

// New creates a new Agent instance
func New(cfg *config.Config, logger *zap.SugaredLogger) (*Agent, error) {
	configUpdater := updater.NewConfigUpdater(cfg, logger)
	return &Agent{
		cfg:            cfg,
		logger:         logger,
		updater:        configUpdater,
		shutdownSignal: make(chan struct{}),
	}, nil
}

// StartAgent starts the agent's core components.
func (a *Agent) StartAgent(ctx context.Context) error {
	if !a.isRunning.CompareAndSwap(false, true) {
		return fmt.Errorf("agent already running")
	}

	setupComplete := false
	defer func() {
		if !setupComplete {
			a.isRunning.Store(false)
			a.logger.Warn("Agent startup failed, reset running state")
		}
	}()

	a.wg.Add(2)
	go func() {
		defer a.wg.Done()
		if err := a.manageCollectorLifecycle(ctx); err != nil {
			a.logger.Errorf("Initial collector run failed: %v", err)
		}
	}()
	go func() {
		defer a.wg.Done()
		a.runConfigUpdateChecker(ctx)
	}()
	a.logger.Info("Agent start sequence initiated.")
	setupComplete = true
	return nil
}

// Shutdown gracefully shuts down the agent
func (a *Agent) Shutdown(ctx context.Context) error {
	if !a.isRunning.CompareAndSwap(true, false) {
		a.logger.Info("Agent shutdown called, but agent is not marked as running.")
		return nil
	}
	close(a.shutdownSignal)
	a.logger.Info("Stopping current collector instance (if any).")
	a.stopCollectorInstance()

	waitCh := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(waitCh)
	}()

	select {
	case <-waitCh:
		a.logger.Info("All agent goroutines completed.")
	case <-ctx.Done():
		a.logger.Errorf("Agent shutdown timed out: %v", ctx.Err())
		return ctx.Err()
	}
	return nil
}

func (a *Agent) manageCollectorLifecycle(ctx context.Context) error {
	if !a.isRunning.Load() {
		a.logger.Info("Agent is shutting down, not starting new collector.")
		return nil
	}

	collector, err := NewCollector(a.cfg)
	if err != nil {
		return fmt.Errorf("failed to create new collector instance: %w", err)
	}
	a.collectorMu.Lock()
	a.collector = collector
	a.collectorMu.Unlock()
	a.logger.Info("Collector instance created. Starting its run loop...")

	runErr := collector.Run(ctx)
	a.logger.Infof("Collector run loop finished. Error: %v", runErr)
	a.collectorMu.Lock()
	if a.collector == collector {
		a.collector = nil
		a.logger.Debug("Collector instance cleared.")
	}
	a.collectorMu.Unlock()

	// ensuring cleanup if this func returns early
	defer func() {
		a.collectorMu.Lock()
		defer a.collectorMu.Unlock()
		if a.collector == collector {
			a.collector = nil
			a.logger.Debug("Collector instance cleared.")
		}
	}()
	return runErr
}

func (a *Agent) stopCollectorInstance() {
	a.collectorMu.Lock()
	collector := a.collector
	a.collector = nil
	a.collectorMu.Unlock()

	if collector != nil {
		a.logger.Info("Initiating shutdown for active collector instance...")
		collector.Shutdown()
		a.logger.Info("Collector shutdown signal sent.")
	}
}

func (a *Agent) UpdateConfig(_ context.Context, newConfig map[string]interface{}) error {
	configYAML, err := yaml.Marshal(newConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal new config to YAML: %w", err)
	}
	tempFile := a.cfg.OtelConfigPath + ".new"
	if err := os.WriteFile(tempFile, configYAML, 0644); err != nil {
		return fmt.Errorf("failed to write new config to temporary file: %w", err)
	}
	if err := os.Rename(tempFile, a.cfg.OtelConfigPath); err != nil {
		return fmt.Errorf("failed to replace config file: %w", err)
	}
	a.logger.Info("Successfully updated collector configuration")
	return nil
}

func (a *Agent) runConfigUpdateChecker(ctx context.Context) {
	if a.cfg.ConfigUpdateURL == "" {
		a.logger.Info("Config update URL not configured, skipping config update checks")
		return
	}
	ticker := time.NewTicker(time.Duration(a.cfg.ConfigCheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := a.performConfigCheck(ctx); err != nil {
				a.logger.Errorf("Periodic config check failed: %v", err)
			}
		case <-a.shutdownSignal:
			a.logger.Info("Config update checker stopping due to shutdown.")
			return
		case <-ctx.Done():
			a.logger.Info("Config update checker stopping due to context cancellation.")
			return
		}
	}
}

func (a *Agent) performConfigCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	a.logger.Info("Checking for configuration updates...")
	restart, newConfig, err := a.updater.CheckForUpdates(ctx)
	if err != nil {
		return fmt.Errorf("updater.CheckForUpdates failed: %w", err)
	}
	if newConfig != nil {
		if err := a.UpdateConfig(ctx, newConfig); err != nil {
			return fmt.Errorf("failed to update config file: %w", err)
		}
	}
	if restart {
		a.logger.Info("Configuration change requires collector restart.")
		if !a.isRunning.Load() {
			a.logger.Info("Agent is shutting down, skipping restart.")
			return nil
		}
		a.stopCollectorInstance()
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			if err := a.manageCollectorLifecycle(ctx); err != nil {
				a.logger.Errorf("Collector restart failed: %v", err)
			} else {
				a.logger.Info("Collector restarted successfully.")
			}
		}()
	}
	return nil
}
