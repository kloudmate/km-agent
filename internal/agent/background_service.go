package agent

import (
	"context"
	"fmt"
	"os"
	"time"

	bgsvc "github.com/kardianos/service"
	"github.com/kloudmate/km-agent/internal/collector"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/httpprovider"
	"go.opentelemetry.io/collector/confmap/provider/httpsprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/otelcol"
	"gopkg.in/yaml.v3"
)

var logger bgsvc.Logger

type KmAgentService struct {
	Collector collector.KmCollector
	Configs   otelcol.CollectorSettings
	Mode      string
	AgentCfg  agentYaml
	Exit      chan struct{}
}

type agentYaml struct {
	Key        string `yaml:"key"`
	debugLevel string `yaml:"debug"`
	Endpoint   string `yaml:"endpoint"`
	Interval   string `yaml:"interval"`
}

// Constructor for the KmAgentService with default configurations
func NewKmAgentService() (*KmAgentService, error) {

	// information about the collector
	info := component.BuildInfo{
		Command:     "kmagent",
		Description: "KloudMate Agent for auto instrumentation",
		Version:     "0.0.1",
	}

	// hardcoded relative path to the default config which will be picked on the initial start.
	uris := []string{HOST_CONFIG_FILE_URI}

	// default configurations for the collector
	set := otelcol.CollectorSettings{
		BuildInfo: info,
		Factories: components,
		ConfigProviderSettings: otelcol.ConfigProviderSettings{
			ResolverSettings: confmap.ResolverSettings{
				URIs: uris,
				ProviderFactories: []confmap.ProviderFactory{
					envprovider.NewFactory(),
					fileprovider.NewFactory(),
					httpprovider.NewFactory(),
					httpsprovider.NewFactory(),
					yamlprovider.NewFactory(),
				},
			},
		},
	}

	kmCollector, err := collector.NewKmCollector(set)
	if err != nil {
		return nil, err
	}

	return &KmAgentService{
		Mode: hostMode,
		// Token:     "",
		AgentCfg:  agentYaml{},
		Configs:   set,
		Collector: *kmCollector,
		Exit:      make(chan struct{}),
	}, nil
}

func (p *KmAgentService) asyncWork() {
	if err := p.Collector.Run(context.Background()); err != nil {
		fmt.Println(fmt.Errorf("error occured while running collector : %s", err.Error()))
	}
}

func (p *KmAgentService) Start(s bgsvc.Service) error {

	fmt.Println(fmt.Sprintf("Running agent on %s mode", p.Mode))
	go p.asyncWork()
	return nil
}

func (p *KmAgentService) Stop(s bgsvc.Service) error {
	fmt.Println("stopping the KloudMate Agent !")
	close(p.Exit)
	return nil
}

// SetToken is used to apply KM_API_KEY on the collector configuration for windows flavoured builds.
func (p *KmAgentService) ApplyAgentConfig() {

	var agentParsedData agentYaml
	p.Collector.SetupConfigurationComponents(context.TODO())

	// reading the default agent configuration and loading them...
	fileData, err := os.ReadFile(AGENT_CONFIG_FILE_URI)
	if err != nil {
		fmt.Printf("failed to read agent configuration : %v \n", err)
	}
	if err := yaml.Unmarshal(fileData, &agentParsedData); err != nil {
		fmt.Printf("failed to parse agent configuration : %v \n", err)
	}

	// if empty and not set on env then use the key from the agent configuration
	if p.AgentCfg.Key == "" {
		p.AgentCfg.Key = agentParsedData.Key
	}

	// if empty and not set on env then use the endpoint from the agent configuration
	if p.AgentCfg.Endpoint == "" {
		p.AgentCfg.Endpoint = agentParsedData.Endpoint
	}

	// if the debug level is not normal then apply it on current configuration.
	if p.AgentCfg.debugLevel != "normal" {
		p.AgentCfg.debugLevel = agentParsedData.debugLevel
	}

	if p.AgentCfg.Interval != "10s" {
		duration, err := time.ParseDuration(p.AgentCfg.Interval)
		if err != nil {
			fmt.Println(fmt.Errorf("error while processing config interval : %v\n", err))
			p.AgentCfg.Interval = "10s"
		}
		// If the duration is less than 10 second then don't apply the interval...
		if duration.Seconds() > 10 {
			p.AgentCfg.Interval = "10s"
		} else {
			p.AgentCfg.Interval = agentParsedData.Interval
		}
	}

	// if found pass then build their uri
	debugUri := fmt.Sprintf("yaml:exporters::debug::verbosity:%s", p.AgentCfg.debugLevel)
	endpointUri := fmt.Sprintf("yaml:exporters::otlphttp::endpoint:%s", p.AgentCfg.Endpoint)
	ApiKeyUri := fmt.Sprintf("yaml:exporters::otlphttp::headers::Authorization:%s", p.AgentCfg.Key)

	// Applying configuration to the agent depending on the mode (i.e - host/ docker)
	if p.Mode == containerMode {
		p.Configs.ConfigProviderSettings.ResolverSettings.URIs =
			[]string{
				DOCKER_CONFIG_FILE_URI,
				ApiKeyUri,
				debugUri,
				endpointUri,
			}
	} else {
		p.Configs.ConfigProviderSettings.ResolverSettings.URIs =
			[]string{
				HOST_CONFIG_FILE_URI,
				ApiKeyUri,
				debugUri,
				endpointUri,
			}
	}

	// reloads the agent configuration

	if p.AgentCfg.Key != agentParsedData.Key {
		// agentParsedData.Key = p.AgentCfg.Key
		p.AgentCfg.Key = agentParsedData.Key
		file, err := os.Create(AGENT_CONFIG_FILE_URI)
		if err != nil {
			logger.Errorf("failed to save agent configuration : %v\n", err)
		}

		enc := yaml.NewEncoder(file)
		enc.SetIndent(2)
		if err = enc.Encode(&agentParsedData); err != nil {
			logger.Errorf("failed to save agent configuration : caused by not able to encode config : %v \n", err)
		}

		defer enc.Close()

		p.Collector.ReloadConfiguration(context.TODO())
		if err != nil {
			logger.Errorf("failed to apply agent configuration : %v \n", err)
		}
	}

}

func (p *KmAgentService) lookupAndUpdateYamlNode(node *yaml.Node, path []string, newVal string, depth int) bool {
	if depth >= len(path) {
		return false
	}
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		return p.lookupAndUpdateYamlNode(node.Content[0], path, newVal, depth)
	}
	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			kNode := node.Content[i]
			valNode := node.Content[i+1]
			if kNode.Value == path[depth] {
				if depth == len(path)-1 {
					valNode.Value = newVal
					return true
				}
				return p.lookupAndUpdateYamlNode(valNode, path, newVal, depth+1)
			}
		}
	}
	return false
}
