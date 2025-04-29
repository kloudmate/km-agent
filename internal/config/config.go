package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

// Config represents the agent configuration
type Config struct {
	// Collector configuration (OpenTelemetry collector config)
	Collector map[string]interface{} `yaml:"collector"`

	// Agent configuration
	Agent struct {
		ExporterEndpoint    string `yaml:"exporter_endpoint"`
		APIKey              string `yaml:"api_key"`
		ConfigCheckInterval int    `yaml:"config_check_interval"`
		DockerMode          bool   `yaml:"docker_mode"`
		ConfigUpdateURL     string `yaml:"config_update_url"`
	} `yaml:"agent"`

	// Path to the configuration file (not stored in the config file itself)
	ConfigPath string `yaml:"-"`
}

// GetDefaultConfigPath returns the default configuration file path based on OS
func GetDefaultConfigPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("ProgramData"), "OtelAgent", "config.yaml")
	} else if runtime.GOOS == "darwin" {
		return "/Library/Application Support/OtelAgent/config.yaml"
	} else {
		// Linux/Unix
		return "/etc/otel-agent/config.yaml"
	}
}

// GetDockerConfigPath returns the configuration path when running in Docker
func GetDockerConfigPath() string {
	return "/etc/otel-agent/docker-config.yaml"
}

// LoadConfig loads the configuration from CLI flags, environment variables, and config file
func LoadConfig(c *cli.Context) (*Config, error) {
	cfg := &Config{}

	// Default config file path based on OS
	configPath := c.String("config")
	if configPath == "" {
		if c.Bool("docker-mode") {
			configPath = GetDockerConfigPath()
		} else {
			configPath = GetDefaultConfigPath()
		}
	}

	// Store the config path
	cfg.ConfigPath = configPath

	// Load config file if exists
	if _, err := os.Stat(configPath); err == nil {
		configData, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(configData, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("error checking config file: %w", err)
	}

	// Override with environment variables and CLI flags
	if endpoint := c.String("exporter-endpoint"); endpoint != "" {
		cfg.Agent.ExporterEndpoint = endpoint
	}

	if apiKey := c.String("api-key"); apiKey != "" {
		cfg.Agent.APIKey = apiKey
	}

	if interval := c.Int("config-check-interval"); interval > 0 {
		cfg.Agent.ConfigCheckInterval = interval
	}

	if c.Bool("docker-mode") {
		cfg.Agent.DockerMode = true
	}

	// Make sure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return cfg, nil
}
