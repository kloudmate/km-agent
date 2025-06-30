package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// Config represents the agent configuration
type Config struct {
	// Collector configuration (OpenTelemetry collector config)
	Collector           map[string]interface{}
	AgentConfigPath     string
	OtelConfigPath      string
	ExporterEndpoint    string
	ConfigUpdateURL     string
	APIKey              string
	ConfigCheckInterval int
	DockerMode          bool
	DockerEndpoint      string
}

// GetDefaultConfigPath returns the default configuration file path based on OS
func GetDefaultConfigPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("ProgramData"), "kmagent", "config.yaml")
	} else if runtime.GOOS == "darwin" {
		return "/Library/Application Support/kmagent/config.yaml"
	} else {
		// Linux/Unix
		return "/etc/kmagent/config.yaml"
	}
}

// GetDockerConfigPath returns the configuration path when running in Docker
func GetDockerConfigPath() string {
	return "/etc/kmagent/config.yaml"
}

// LoadConfig loads the configuration from CLI flags, environment variables, and config file
// TODO It should load config from server as well
func (c *Config) LoadConfig() error {

	os.Setenv("KM_COLLECTOR_ENDPOINT", c.ExporterEndpoint)
	os.Setenv("KM_API_KEY", c.APIKey)

	// Default config file path based on OS
	configPath := c.OtelConfigPath
	if configPath == "" {
		if c.DockerMode {
			configPath = GetDockerConfigPath()
		} else {
			configPath = GetDefaultConfigPath()
		}
	}

	// Store the config path
	c.OtelConfigPath = configPath

	// Load config file if exists
	if _, err := os.Stat(configPath); err == nil {
		configData, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(configData, c); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error checking config file: %w", err)
	}

	// Make sure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return nil
}

func (c *Config) Hostname() string {
	n, e := os.Hostname()
	if e != nil {
		n = ""
	}
	return n
}

// UpdateConfigFile TODO it should update the config json from api with relevant details and save to fs
func UpdateConfigFile() {

}
