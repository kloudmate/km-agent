package agent

import (
	"context"
	"fmt"
	"os"
	"strings"

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
	Token     string
	Exit      chan struct{}
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
	uris := []string{CONFIG_FILE_URI}

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
		Mode:      "host",
		Token:     "",
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
	if p.Mode == "docker" {
		p.Configs.ConfigProviderSettings.ResolverSettings.URIs = []string{DOCKER_CONFIG_FILE_URI}
		col, err := collector.NewKmCollector(p.Configs)
		if err != nil {
			return err
		}
		p.Collector = *col
	}
	fmt.Println(fmt.Sprintf("Running agent on %s mode", p.Mode))
	go p.asyncWork()
	return nil
}

func (p *KmAgentService) Stop(s bgsvc.Service) error {
	fmt.Println("stopping the KloudMate Agent !")
	close(p.Exit)
	return nil
}

func (p *KmAgentService) SetToken() {
	var apiKey string
	keyFromEnv := os.Getenv("KM_API_KEY")
	if p.Token == "" && keyFromEnv == "" {
		return
	}
	if keyFromEnv != "" {
		apiKey = keyFromEnv
	} else {
		apiKey = p.Token
	}

	var configFile string
	if p.Mode == "host" {
		configFile = CONFIG_FILE_URI
	} else {
		configFile = DOCKER_CONFIG_FILE_URI
	}

	fileData, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Printf("failed to set Token : caused by not able to read file : %v \n", err)
	}

	var parsedData yaml.Node
	if err := yaml.Unmarshal(fileData, &parsedData); err != nil {
		fmt.Printf("failed to set Token : caused by not able to unmarshal config file : %v \n", err)
	}

	nPaths := strings.Split("exporters.otlphttp.headers.Authorization", ".")
	isModified := p.lookupAndUpdateYamlNode(&parsedData, nPaths, apiKey, 0)

	if isModified {
		// creating temp file to store modified configuration
		tmpFile, err := os.CreateTemp("", "conf-tmp-*.yaml")
		if err != nil {
			fmt.Printf("failed to set Token : caused by not able to create temp file : %s \n", err.Error())
		}
		tmpFilePath := tmpFile.Name()

		defer os.Remove(tmpFilePath)
		defer tmpFile.Close()

		enc := yaml.NewEncoder(tmpFile)
		enc.SetIndent(2)
		if err = enc.Encode(&parsedData); err != nil {
			fmt.Printf("failed to set Token : caused by not able to encode modified config : %s \n", err.Error())
		}
		enc.Close()
		tmpFile.Close()

		if err = os.Rename(tmpFilePath, configFile); err != nil {
			fmt.Printf("failed to set Token : caused by not able to rename temp file to original config : %s \n", err.Error())
		}
	} else {
		fmt.Printf("failed to set Token : caused by not able to locate the kloudmate based node in config file : %v \n", err)
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
