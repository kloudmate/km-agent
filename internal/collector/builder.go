package collector

import (
	"context"
	"fmt"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kloudmate/km-agent/internal/config"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/otelcol"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// Collector represents the OpenTelemetry collector
type Collector struct {
	cfg        *config.Config
	logger     *zap.SugaredLogger
	otelCol    *otelcol.Collector
	mu         sync.Mutex
	isRunning  bool
	configPath string
}

// NewCollector creates a new Collector instance
func NewCollector(cfg *config.Config, logger *zap.SugaredLogger) (*Collector, error) {
	// Determine config path
	configPath := cfg.ConfigPath
	if configPath == "" {
		if cfg.Agent.DockerMode {
			configPath = config.GetDockerConfigPath()
		} else {
			configPath = config.GetDefaultConfigPath()
		}
	}

	// Make sure parent directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory %s: %w", configDir, err)
	}

	return &Collector{
		cfg:        cfg,
		logger:     logger,
		configPath: configPath,
	}, nil
}

// Start starts the collector
func (c *Collector) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isRunning {
		return fmt.Errorf("collector is already running")
	}

	// Create collector settings
	set := otelcol.CollectorSettings{
		BuildInfo: component.BuildInfo{
			Command:     "kmagent",
			Description: "KloudMate Agent",
			Version:     "1.0.0",
		},
		Factories:               components,
		DisableGracefulShutdown: false,
		LoggingOptions:          []zap.Option{},
	}

	// Check if config file exists, if not write a default config
	if _, err := os.Stat(c.configPath); os.IsNotExist(err) {
		c.logger.Error("Config file not found")
		//c.logger.Infof("Config file %s not found, creating default config", c.configPath)
		//if err := c.writeDefaultConfig(); err != nil {
		//	return fmt.Errorf("failed to write default config: %w", err)
		//}
	}
	// Create config provider
	set.ConfigProviderSettings = otelcol.ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			DefaultScheme: "env",
			URIs:          []string{c.configPath},
			ProviderFactories: []confmap.ProviderFactory{
				envprovider.NewFactory(),
				fileprovider.NewFactory(),
				yamlprovider.NewFactory(),
			},
		},
	}

	// Create the collector
	otelCol, err := otelcol.NewCollector(set)
	if err != nil {
		return fmt.Errorf("failed to create collector: %w", err)
	}

	// Save the collector
	c.otelCol = otelCol
	c.isRunning = true

	// Start the collector in a goroutine
	go func() {
		if err := otelCol.Run(ctx); err != nil {
			c.logger.Errorf("Collector stopped with error: %v", err)
		}

		c.mu.Lock()
		c.isRunning = false
		c.mu.Unlock()
	}()

	return nil
}

// Shutdown shuts down the collector
func (c *Collector) Shutdown(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isRunning || c.otelCol == nil {
		return nil
	}

	// Cancel the collector context
	c.otelCol.Shutdown()
	c.isRunning = false
	return nil
}

// UpdateConfig updates the collector configuration with a new config from remote
func (c *Collector) UpdateConfig(ctx context.Context, newConfig map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Convert config to YAML and write to file
	configYAML, err := yaml.Marshal(newConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal new config to YAML: %w", err)
	}

	// Create a temporary file
	tempFile := c.configPath + ".new"
	if err := os.WriteFile(tempFile, configYAML, 0644); err != nil {
		return fmt.Errorf("failed to write new config to temporary file: %w", err)
	}

	// Rename the temporary file to the actual config file (atomic operation)
	if err := os.Rename(tempFile, c.configPath); err != nil {
		return fmt.Errorf("failed to replace config file: %w", err)
	}

	c.logger.Info("Successfully updated collector configuration")
	return nil
}

// writeDefaultConfig writes a default configuration to the config file
func (c *Collector) writeDefaultConfig() error {
	// Create a default config based on whether we're in Docker mode
	var defaultConfig string
	if c.cfg.Agent.DockerMode {
		defaultConfig = `receivers:
  docker_stats:
    collection_interval: 30s
    timeout: 20s
    endpoint: unix:///var/run/docker.sock

processors:
  resource:
    attributes:
      - key: service.name
        value: custom-otel-agent
        action: upsert

exporters:
  otlp:
    endpoint: localhost:4317
    tls:
      insecure: true

service:
  pipelines:
    metrics:
      receivers: [docker_stats]
      processors: [resource]
      exporters: [otlp]
`
	} else {
		defaultConfig = `receivers:
  hostmetrics:
    collection_interval: 30s
    scrapers:
      cpu:
      memory:
      disk:
      network:
  docker_stats:
    collection_interval: 30s
    timeout: 20s

processors:
  resource:
    attributes:
      - key: service.name
        value: custom-otel-agent
        action: upsert

exporters:
  otlp:
    endpoint: localhost:4317
    tls:
      insecure: true

service:
  pipelines:
    metrics:
      receivers: [hostmetrics, docker_stats]
      processors: [resource]
      exporters: [otlp]
`
	}

	// Update with any CLI/env settings
	if c.cfg.Agent.ExporterEndpoint != "" {
		// In reality, you'd parse and modify the YAML properly
		// This is just a simple string replacement for demonstration
		defaultConfig = strings.Replace(
			defaultConfig,
			"endpoint: localhost:4317",
			fmt.Sprintf("endpoint: %s", c.cfg.Agent.ExporterEndpoint),
			1,
		)
	}

	// Write the configuration to file
	return os.WriteFile(c.configPath, []byte(defaultConfig), 0644)
}
