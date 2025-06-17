package kube

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DefaultAgentConfigPath = "/var/kloudmate/agent-config.yaml"

	EnvAPIKey              = "KM_API_KEY"
	EnvAgentConfig         = "KM_AGENT_CONFIG"
	EnvExporterEndpoint    = "KM_COLLECTOR_ENDPOINT"
	EnvConfigCheckInterval = "KM_CONFIG_CHECK_INTERVAL"
	EnvUpdateEndpoint      = "KM_UPDATE_ENDPOINT"
)

type KubeAgentConfig struct {
	AgentConfigPath     string
	ExporterEndpoint    string
	ConfigUpdateURL     string
	APIKey              string `yaml:"apikey"`
	ConfigCheckInterval time.Duration

	//k8s monitoring configs
	Monitoring struct {
		ClusterName        string `yaml:"cluster_name"`
		CollectionInterval string `yaml:"collection_interval"`

		Cluster struct {
			Enabled bool     `yaml:"enabled"`
			Metrics []string `yaml:"metrics"`
		} `yaml:"cluster"`

		Nodes struct {
			Enabled       bool     `yaml:"enabled"`
			MonitorAll    bool     `yaml:"monitor_all"`
			SpecificNodes []string `yaml:"specific_nodes"`
			Metrics       []string `yaml:"metrics"`
		} `yaml:"nodes"`

		Pods struct {
			Enabled              bool `yaml:"enabled"`
			MonitorAllNamespaces bool `yaml:"monitor_all_namespaces"`

			Namespaces struct {
				Include []string `yaml:"include"`
				Exclude []string `yaml:"exclude"`
			} `yaml:"namespaces"`

			SpecificPods []struct {
				Name      string `yaml:"name"`
				Namespace string `yaml:"namespace"`
			} `yaml:"specific_pods"`

			Metrics []string `yaml:"metrics"`
		} `yaml:"pods"`

		NamedResources struct {
			Enabled bool `yaml:"enabled"`

			Deployments []struct {
				Name      string `yaml:"name"`
				Namespace string `yaml:"namespace"`
			} `yaml:"deployments"`

			Services []struct {
				Name      string `yaml:"name"`
				Namespace string `yaml:"namespace"`
			} `yaml:"services"`

			ConfigMaps []struct {
				Name      string `yaml:"name"`
				Namespace string `yaml:"namespace"`
			} `yaml:"configmaps"`

			Secrets []struct {
				Name      string `yaml:"name"`
				Namespace string `yaml:"namespace"`
			} `yaml:"secrets"`

			PersistentVolumes []struct {
				Name string `yaml:"name"`
			} `yaml:"persistent_volumes"`

			Metrics []string `yaml:"metrics"`
		} `yaml:"named_resources"`

		Logs struct {
			Enabled bool     `yaml:"enabled"`
			Sources []string `yaml:"sources"`
		} `yaml:"logs"`
	} `yaml:"monitoring"`
}

func writeTempOtelConfig(yamlBytes []byte) (string, error) {
	tmpFile, err := os.CreateTemp("", "otel-config-*.yaml")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	_, err = tmpFile.Write(yamlBytes)
	if err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

// LoadKubeAgentConfig loads and parses the agent config from the default path for kube agent.
func LoadKubeAgentConfig() (*KubeAgentConfig, error) {
	agentConfig := ""

	// Set AgentConfig from environment variable if available
	if envAgentConfig := os.Getenv(EnvAgentConfig); envAgentConfig != "" {
		agentConfig = envAgentConfig
	} else {
		agentConfig = DefaultAgentConfigPath
	}

	// Check if file exists and is readable
	info, err := os.Stat(agentConfig)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("agent config file does not exist at %s", agentConfig)
		}
		return nil, fmt.Errorf("error accessing agent config: %w", err)
	}
	if info.IsDir() {
		return nil, errors.New("agent config path is a directory, expected a file")
	}

	// Read file contents
	data, err := os.ReadFile(agentConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to read agent config file: %w", err)
	}

	// Parse YAML into struct
	var cfg KubeAgentConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse agent config YAML: %w", err)
	}

	// Override API key from environment variable if available
	if envAPIKey := os.Getenv(EnvAPIKey); envAPIKey != "" {
		cfg.APIKey = envAPIKey
	}

	// Override ExporterEndpoint from environment variable if available
	if envExporterEndpoint := os.Getenv(EnvExporterEndpoint); envExporterEndpoint != "" {
		cfg.ExporterEndpoint = envExporterEndpoint
	}

	// Override UpdateEndpoint from environment variable if available
	if envUpdateEndpoint := os.Getenv(EnvUpdateEndpoint); envUpdateEndpoint != "" {
		cfg.ConfigUpdateURL = envUpdateEndpoint
	}

	// Override Config check interval from environment variable if available
	if envConfigCheckInterval := os.Getenv(EnvConfigCheckInterval); envConfigCheckInterval != "" {
		duration, err := time.ParseDuration(envConfigCheckInterval)
		if err != nil {
			fmt.Errorf("failed to parse config check interval from env falling back to 10s (deafault): %w", err)
			duration = time.Duration(time.Second * 10)
		}
		cfg.ConfigCheckInterval = duration
	}

	return &cfg, nil
}
