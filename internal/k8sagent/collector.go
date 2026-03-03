package k8sagent

import (
	"context"
	"fmt"
	"time"

	"github.com/kloudmate/km-agent/internal/shared"
	"go.opentelemetry.io/collector/otelcol"
)

func (a *K8sAgent) startInternalCollector() error {
	a.collectorMu.Lock()
	defer a.collectorMu.Unlock()

	a.Logger.Info("starting collector instance")

	collectorSettings := shared.CollectorInfoFactory(a.otelConfigPath())
	if a.Cfg.DeploymentMode == "DEPLOYMENT" {
		factories, err := collectorSettings.Factories()
		if err == nil {
			// eBPF receiver cannot run in deployment mode (needs host access)
			for typeName := range factories.Receivers {
				if typeName.String() == "ebpfreceiver" {
					delete(factories.Receivers, typeName)
				}
			}
			collectorSettings.Factories = func() (otelcol.Factories, error) {
				return factories, nil
			}
		}
	}
	// Create a context for this collector instance
	a.collectorCtx, a.collectorCancel = context.WithCancel(context.Background())

	collector, err := otelcol.NewCollector(collectorSettings)
	if err != nil {
		a.collectorCancel()
		return fmt.Errorf("failed to create new collector: %w", err)
	}
	a.Collector = collector

	// Start the collector in a separate goroutine.
	a.wg.Add(1)
	go func(col *otelcol.Collector, ctx context.Context) {
		defer a.wg.Done()

		a.Logger.Infow("collector starting",
			"configPath", a.otelConfigPath(),
			"deploymentMode", a.Cfg.DeploymentMode,
		)

		runErr := col.Run(ctx)

		a.collectorMu.Lock()
		if a.Collector == col {
			a.Collector = nil
		}
		a.collectorMu.Unlock()

		if runErr != nil {
			a.Logger.Errorw("collector exited with error", "error", runErr)
		} else {
			a.Logger.Info("collector exited normally")
		}
	}(a.Collector, a.collectorCtx)

	a.Logger.Info("collector instance started")
	return nil
}

// stopInternalCollector gracefully stops the currently running collector instance.
func (a *K8sAgent) stopInternalCollector() {
	a.collectorMu.Lock()
	defer a.collectorMu.Unlock()

	if a.Collector == nil {
		a.Logger.Debug("no active collector instance to stop")
		return
	}

	a.Logger.Info("stopping collector instance")

	// Signal the collector's context to cancel its operations
	if a.collectorCancel != nil {
		a.collectorCancel()
	}

	// Create a context with a timeout for the shutdown operation itself
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Use a channel to know when the collector's Shutdown method completes
	done := make(chan struct{})
	go func() {
		a.Collector.Shutdown()
		close(done)
	}()

	select {
	case <-done:
		a.Logger.Info("collector instance stopped successfully")
	case <-shutdownCtx.Done():
		a.Logger.Warnw("collector shutdown timed out", "timeout", "10s", "error", shutdownCtx.Err())
	}

	a.Collector = nil
	a.collectorCancel = nil
}
