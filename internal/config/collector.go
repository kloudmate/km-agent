package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type CollectorConfig struct {
	// OpenTelemetry Collector Configuration
	Receivers  map[string]interface{} `mapstructure:"receivers"`
	Processors map[string]interface{} `mapstructure:"processors"`
	Exporters  map[string]interface{} `mapstructure:"exporters"`
	Extensions map[string]interface{} `mapstructure:"extensions"`
	Service    ServiceConfig          `mapstructure:"service"`
}

type ServiceConfig struct {
	Extensions []string `mapstructure:"extensions"`
	Pipelines  struct {
		Traces  []string `mapstructure:"traces"`
		Metrics []string `mapstructure:"metrics"`
		Logs    []string `mapstructure:"logs"`
	} `mapstructure:"pipelines"`
}

func LoadCollectorConfig() (*CollectorConfig, error) {
	collectorCfg := &CollectorConfig{}

	// Use Viper to load collector-specific configuration
	viper.SetConfigName("collector")
	viper.SetConfigType("yaml")

	// Add potential config locations
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.kmagent")
	viper.AddConfigPath("/etc/kmagent")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading collector config: %w", err)
	}

	// Unmarshal into CollectorConfig
	if err := viper.Unmarshal(collectorCfg); err != nil {
		return nil, fmt.Errorf("unable to decode collector config: %w", err)
	}

	return collectorCfg, nil
}
