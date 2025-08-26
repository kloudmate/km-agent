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

	a.Logger.Infoln("Starting actual collector instance with new configuration...")

	collectorSettings := shared.CollectorInfoFactory(a.otelConfigPath())

	// Create a context for this collector instance
	a.collectorCtx, a.collectorCancel = context.WithCancel(context.Background())

	collector, err := otelcol.NewCollector(collectorSettings)
	if err != nil {
		a.collectorCancel()
		return fmt.Errorf("failed to create new collector: %w", err)
	}
	a.Collector = collector

	// Start the collector in a separate goroutine.
	// The collector's Start method is blocking until the collector is ready or fails.
	a.wg.Add(1)
	go func(col *otelcol.Collector, ctx context.Context) {
		defer func() {
			a.wg.Done()
			if err != nil {
				a.Logger.Infoln("Collector exited with error: %v", err)

			}
		}()

		a.Logger.Infof("Collector: Starting with configMap mounted in path:  [%s] \n", a.otelConfigPath())
		err = col.Run(ctx)
		if err != nil {
			a.Logger.Infoln("Collector exited with error: %v", err)

		} else {
			a.Logger.Infoln("Collector exited normally.")
		}
		a.collectorMu.Lock()
		defer func() {
			if err != nil {
				a.Logger.Infoln("Collector exited with error: %v", err)

			}
			a.collectorMu.Unlock()
		}()

		if a.Collector == col {
			a.Collector = nil
		}

	}(a.Collector, a.collectorCtx)

	a.Logger.Infoln("Collector instance started.")
	return nil
}

// stopInternalCollector gracefully stops the currently running collector instance.
func (a *K8sAgent) stopInternalCollector() {
	a.collectorMu.Lock()
	defer a.collectorMu.Unlock()

	a.Logger.Infoln("Attempting to stop collector instance gracefully...")

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
		a.Logger.Infoln("Collector instance stopped successfully.")
	case <-shutdownCtx.Done():
		a.Logger.Infoln("Collector shutdown timed out after 10 seconds: %v", shutdownCtx.Err())
	}

	a.Collector = nil
	a.collectorCancel = nil
}
