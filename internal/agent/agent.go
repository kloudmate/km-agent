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
	collectorError string
	version        string
}

type Option func(a *Agent)

func WithVersion(v string) Option {
	return func(a *Agent) {
		a.version = v
	}
}

// New creates a new Agent instance
func New(cfg *config.Config, logger *zap.SugaredLogger, opts ...Option) (*Agent, error) {
	configUpdater := updater.NewConfigUpdater(cfg, logger)
	a := Agent{
		cfg:            cfg,
		logger:         logger,
		updater:        configUpdater,
		shutdownSignal: make(chan struct{}),
	}

	for _, o := range opts {
		o(&a)
	}

	return &a, nil
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
	a.logger.Info("agent start sequence initiated")
	setupComplete = true
	return nil
}

// Shutdown gracefully shuts down the agent
func (a *Agent) Shutdown(ctx context.Context) error {
	if !a.isRunning.CompareAndSwap(true, false) {
		a.logger.Debug("shutdown called but agent is not running")
		return nil
	}
	close(a.shutdownSignal)
	a.logger.Info("stopping collector instance")
	a.stopCollectorInstance()

	waitCh := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(waitCh)
	}()

	select {
	case <-waitCh:
		a.logger.Info("all agent goroutines completed")
	case <-ctx.Done():
		a.logger.Errorf("Agent shutdown timed out: %v", ctx.Err())
		return ctx.Err()
	}
	return nil
}

func (a *Agent) manageCollectorLifecycle(ctx context.Context) error {
	// Initial check to exit early and avoid unnecessary work.
	if !a.isRunning.Load() {
		a.logger.Info("agent shutting down, skipping collector start")
		return nil
	}

	// Create the collector instance.
	collector, err := NewCollector(a.cfg)
	if err != nil {
		return fmt.Errorf("failed to create new collector instance: %w", err)
	}

	// This deferred function will run when manageCollectorLifecycle exits for any reason.
	// It ensures the collector reference is cleaned up safely.
	defer func() {
		a.collectorMu.Lock()
		defer a.collectorMu.Unlock()
		// This check is crucial: only clear the reference if it's the one we managed.
		// This prevents a race condition if another collector was started in the meantime.
		if a.collector == collector {
			a.collector = nil
			a.logger.Debug("collector instance cleared")
		}
	}()

	// Atomically check the running state and assign the new collector.
	a.collectorMu.Lock()
	if !a.isRunning.Load() {
		// The agent was shut down between the initial check and now. Abort.
		a.collectorMu.Unlock()
		a.logger.Info("agent shutdown initiated, aborting collector start")
		return nil // Or a specific error if desired, like context.Canceled
	}
	a.collector = collector
	a.collectorMu.Unlock()

	a.logger.Info("collector instance created, starting run loop")
	runErr := collector.Run(ctx)
	if runErr != nil {
		a.collectorError = runErr.Error()
		a.logger.Errorw("collector run loop exited with error", "error", runErr)
	} else {
		a.collectorError = ""
		a.logger.Info("collector run loop exited normally")
	}

	return runErr
}

func (a *Agent) stopCollectorInstance() {
	a.collectorMu.Lock()
	collector := a.collector
	a.collector = nil
	a.collectorMu.Unlock()

	if collector != nil {
		a.logger.Info("shutting down active collector instance")
		collector.Shutdown()
		a.logger.Info("collector shutdown complete")
	}
}

// UpdateConfig takes new config and create new otel config file and update existing config file.
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
	a.logger.Infow("collector configuration updated", "configPath", a.cfg.OtelConfigPath)
	return nil
}

// runConfigUpdateChecker run ticker for performConfigCheck
func (a *Agent) runConfigUpdateChecker(ctx context.Context) {
	if a.cfg.ConfigUpdateURL == "" {
		a.logger.Debug("config update URL not configured, skipping update checks")
		return
	}
	if a.cfg.ConfigCheckInterval <= 0 {
		a.logger.Debug("config check interval not set, skipping update checks")
		return
	}
	a.logger.Infow("config update checker started",
		"updateURL", a.cfg.ConfigUpdateURL,
		"intervalSeconds", a.cfg.ConfigCheckInterval,
	)
	ticker := time.NewTicker(time.Duration(a.cfg.ConfigCheckInterval) * time.Second)
	defer ticker.Stop()

	// trigger the very first config check
	if err := a.performConfigCheck(ctx); err != nil {
		a.logger.Errorf("Periodic config check failed: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := a.performConfigCheck(ctx); err != nil {
				a.logger.Errorf("Periodic config check failed: %v", err)
			}
		case <-a.shutdownSignal:
			a.logger.Info("config update checker stopping")
			return
		case <-ctx.Done():
			a.logger.Info("config update checker stopping")
			return
		}
	}
}

// performConfigCheck checks remote server for new config and restart collector if required
func (a *Agent) performConfigCheck(agentCtx context.Context) error {
	ctx, cancel := context.WithTimeout(agentCtx, 10*time.Second)
	defer cancel()

	a.logger.Debug("checking for configuration updates")

	a.collectorMu.Lock()
	params := updater.UpdateCheckerParams{
		Version: a.version,
	}
	if a.collector != nil {
		params.CollectorStatus = "Running"
	} else {
		params.CollectorStatus = "Stopped"
		params.CollectorLastError = a.collectorError // Safe to read now
	}
	a.collectorMu.Unlock()

	if a.isRunning.Load() {
		params.AgentStatus = "Running"
	} else {
		params.AgentStatus = "Stopped"
	}

	a.logger.Debugf("Checking for updates with params: %+v", params)

	restart, newConfig, err := a.updater.CheckForUpdates(ctx, params)
	if err != nil {
		return fmt.Errorf("updater.CheckForUpdates failed: %w", err)
	}
	if newConfig != nil && restart {
		if err := a.UpdateConfig(ctx, newConfig); err != nil {
			a.collectorError = err.Error()
			return fmt.Errorf("failed to update config file: %w", err)
		}
		a.logger.Info("configuration changed, restarting collector")
		if !a.isRunning.Load() {
			a.logger.Info("agent shutting down, skipping restart")
			return nil
		}

		a.stopCollectorInstance()
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			if err := a.manageCollectorLifecycle(agentCtx); err != nil {
				a.collectorError = err.Error()
			} else {
				a.logger.Info("collector restarted successfully")
				a.collectorError = ""
			}
		}()
	} else {
		a.logger.Debug("no configuration change detected")
	}
	return nil
}
