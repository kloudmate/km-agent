package collector

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	bgsvc "github.com/kardianos/service"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/kloudmate/km-agent/internal/collector/grpclog"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/service"
)

var log bgsvc.Logger

// KmCollector is a wrapper around opentelemetry's collector
type KmCollector struct {
	set otelcol.CollectorSettings

	configProvider otelcol.ConfigProvider

	serviceConfig *service.Config
	service       *service.Service
	state         *atomic.Int64

	// shutdownChan is used to terminate the collector.
	shutdownChan chan struct{}
	// signalsChannel is used to receive termination signals from the OS.
	signalsChannel chan os.Signal
	// asyncErrorChannel is used to signal a fatal error from any component.
	asyncErrorChannel          chan error
	bc                         *bufferedCore
	updateConfigProviderLogger func(core zapcore.Core)
}

// State defines Collector's state.
type State int

const (
	StateStarting State = iota
	StateRunning
	StateClosing
	StateClosed
)

func (s State) String() string {
	switch s {
	case StateStarting:
		return "Starting"
	case StateRunning:
		return "Running"
	case StateClosing:
		return "Closing"
	case StateClosed:
		return "Closed"
	}
	return "UNKNOWN"
}

// (Internal note) Collector Lifecycle:
// - New constructs a new Collector.
// - Run starts the collector.
// - Run calls setupConfigurationComponents to handle configuration.
//   If configuration parser fails, collector's config can be reloaded.
//   Collector can be shutdown if parser gets a shutdown error.
// - Run runs runAndWaitForShutdownEvent and waits for a shutdown event.
//   SIGINT and SIGTERM, errors, and (*KmCollector).Shutdown can trigger the shutdown events.
// - Upon shutdown, pipelines are notified, then pipelines and extensions are shut down.
// - Users can call (*KmCollector).Shutdown anytime to shut down the collector.

// Collector represents a server providing the OpenTelemetry Collector service.

// NewCollector creates and returns a new instance of Collector.
func NewKmCollector(set otelcol.CollectorSettings) (*KmCollector, error) {
	bc := newBufferedCore(zapcore.DebugLevel)
	cc := &collectorCore{core: bc}
	options := append([]zap.Option{zap.WithCaller(true)}, set.LoggingOptions...)
	logger := zap.New(cc, options...)
	set.ConfigProviderSettings.ResolverSettings.ProviderSettings = confmap.ProviderSettings{Logger: logger}
	set.ConfigProviderSettings.ResolverSettings.ConverterSettings = confmap.ConverterSettings{Logger: logger}

	configProvider, err := otelcol.NewConfigProvider(set.ConfigProviderSettings)
	if err != nil {
		return nil, err
	}

	state := new(atomic.Int64)
	state.Store(int64(StateStarting))
	return &KmCollector{
		set:          set,
		state:        state,
		shutdownChan: make(chan struct{}),
		// Per signal.Notify documentation, a size of the channel equaled with
		// the number of signals getting notified on is recommended.
		signalsChannel:             make(chan os.Signal, 3),
		asyncErrorChannel:          make(chan error),
		configProvider:             configProvider,
		bc:                         bc,
		updateConfigProviderLogger: cc.SetCore,
	}, nil
}

// GetState returns current state of the collector server.
func (col *KmCollector) GetState() State {
	return State(col.state.Load())
}

// Shutdown shuts down the collector server.
func (col *KmCollector) Shutdown() {
	// Only shutdown if we're in a Running or Starting State else noop
	state := col.GetState()
	if state == StateRunning || state == StateStarting {
		defer func() {
			recover() // nolint:errcheck
		}()
		close(col.shutdownChan)
	}
}

// setupConfigurationComponents loads the config, creates the graph, and starts the components. If all the steps succeeds it
// sets the col.service with the service currently running.
func (col *KmCollector) setupConfigurationComponents(ctx context.Context) error {
	col.setCollectorState(StateStarting)

	factories, err := col.set.Factories()
	if err != nil {
		return fmt.Errorf("failed to initialize factories: %w", err)
	}
	cfg, err := col.configProvider.Get(ctx, factories)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	if err = cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	col.serviceConfig = &cfg.Service

	conf := confmap.New()

	if err = conf.Marshal(cfg); err != nil {
		return fmt.Errorf("could not marshal configuration: %w", err)
	}

	col.service, err = service.New(ctx, service.Settings{
		BuildInfo:     col.set.BuildInfo,
		CollectorConf: conf,

		ReceiversConfigs:    cfg.Receivers,
		ReceiversFactories:  factories.Receivers,
		ProcessorsConfigs:   cfg.Processors,
		ProcessorsFactories: factories.Processors,
		ExportersConfigs:    cfg.Exporters,
		ExportersFactories:  factories.Exporters,
		ConnectorsConfigs:   cfg.Connectors,
		ConnectorsFactories: factories.Connectors,
		ExtensionsConfigs:   cfg.Extensions,
		ExtensionsFactories: factories.Extensions,

		ModuleInfo: extension.ModuleInfo{
			Receiver:  factories.ReceiverModules,
			Processor: factories.ProcessorModules,
			Exporter:  factories.ExporterModules,
			Extension: factories.ExtensionModules,
			Connector: factories.ConnectorModules,
		},
		AsyncErrorChannel: col.asyncErrorChannel,
		LoggingOptions:    col.set.LoggingOptions,
	}, cfg.Service)
	if err != nil {
		return err
	}
	if col.updateConfigProviderLogger != nil {
		col.updateConfigProviderLogger(col.service.Logger().Core())
	}
	if col.bc != nil {
		x := col.bc.TakeLogs()
		for _, log := range x {
			ce := col.service.Logger().Core().Check(log.Entry, nil)
			if ce != nil {
				ce.Write(log.Context...)
			}
		}
	}

	if !col.set.SkipSettingGRPCLogger {
		grpclog.SetLogger(col.service.Logger(), cfg.Service.Telemetry.Logs.Level)
	}

	if err = col.service.Start(ctx); err != nil {
		return multierr.Combine(err, col.service.Shutdown(ctx))
	}
	col.setCollectorState(StateRunning)

	return nil
}

func (col *KmCollector) reloadConfiguration(ctx context.Context) error {
	col.service.Logger().Warn("Config updated, restart service")
	col.setCollectorState(StateClosing)

	if err := col.service.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown the retiring config: %w", err)
	}

	if err := col.setupConfigurationComponents(ctx); err != nil {
		return fmt.Errorf("failed to setup configuration components: %w", err)
	}

	return nil
}

func (col *KmCollector) DryRun(ctx context.Context) error {
	factories, err := col.set.Factories()
	if err != nil {
		return fmt.Errorf("failed to initialize factories: %w", err)
	}
	cfg, err := col.configProvider.Get(ctx, factories)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	return cfg.Validate()
}

func newFallbackLogger(options []zap.Option) (*zap.Logger, error) {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeTime = zapcore.ISO8601TimeEncoder
	zapCfg := &zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Encoding:         "console",
		EncoderConfig:    ec,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
	return zapCfg.Build(options...)
}

// Run starts the collector according to the given configuration, and waits for it to complete.
// Consecutive calls to Run are not allowed, Run shouldn't be called once a collector is shut down.
// Sets up the control logic for config reloading and shutdown.
func (col *KmCollector) Run(ctx context.Context) error {
	// setupConfigurationComponents is the "main" function responsible for startup
	if err := col.setupConfigurationComponents(ctx); err != nil {
		col.setCollectorState(StateClosed)
		logger, loggerErr := newFallbackLogger(col.set.LoggingOptions)
		if loggerErr != nil {
			return errors.Join(err, fmt.Errorf("unable to create fallback logger: %w", loggerErr))
		}

		if col.bc != nil {
			x := col.bc.TakeLogs()
			for _, log := range x {
				ce := logger.Core().Check(log.Entry, nil)
				if ce != nil {
					ce.Write(log.Context...)
				}
			}
		}

		return err
	}

	// Always notify with SIGHUP for configuration reloading.
	signal.Notify(col.signalsChannel, syscall.SIGHUP)
	defer signal.Stop(col.signalsChannel)

	// Only notify with SIGTERM and SIGINT if graceful shutdown is enabled.
	if !col.set.DisableGracefulShutdown {
		signal.Notify(col.signalsChannel, os.Interrupt, syscall.SIGTERM)
	}

	// Control loop: selects between channels for various interrupts - when this loop is broken, the collector exits.
	// If a configuration reload fails, we return without waiting for graceful shutdown.

	go func() {
		for {
			time.Sleep(time.Second * 5)
			col.service.Logger().Warn("performing configuration reload")
			// if err := col.reloadConfiguration(ctx); err != nil {
			// 	col.service.Logger().Warn("performing configuration reload failed")
			// }
		}

	}()
LOOP:
	for {
		select {
		case err := <-col.configProvider.Watch():
			if err != nil {
				col.service.Logger().Error("Config watch failed", zap.Error(err))
				break LOOP
			}
			if err = col.reloadConfiguration(ctx); err != nil {
				return err
			}
		case err := <-col.asyncErrorChannel:
			col.service.Logger().Error("Asynchronous error received, terminating process", zap.Error(err))
			break LOOP
		case s := <-col.signalsChannel:
			col.service.Logger().Info("Received signal from OS", zap.String("signal", s.String()))
			if s != syscall.SIGHUP {
				break LOOP
			}
			if err := col.reloadConfiguration(ctx); err != nil {
				return err
			}
		case <-col.shutdownChan:
			col.service.Logger().Info("Received shutdown request")
			break LOOP
		case <-ctx.Done():
			col.service.Logger().Info("Context done, terminating process", zap.Error(ctx.Err()))
			// Call shutdown with background context as the passed in context has been canceled
			return col.shutdown(context.Background())
		}
	}
	return col.shutdown(ctx)
}

func (col *KmCollector) shutdown(ctx context.Context) error {
	col.setCollectorState(StateClosing)

	// Accumulate errors and proceed with shutting down remaining components.
	var errs error

	if err := col.configProvider.Shutdown(ctx); err != nil {
		errs = multierr.Append(errs, fmt.Errorf("failed to shutdown config provider: %w", err))
	}

	// shutdown service
	if err := col.service.Shutdown(ctx); err != nil {
		errs = multierr.Append(errs, fmt.Errorf("failed to shutdown service after error: %w", err))
	}

	col.setCollectorState(StateClosed)

	return errs
}

// setCollectorState provides current state of the collector
func (col *KmCollector) setCollectorState(state State) {
	col.state.Store(int64(state))
}
