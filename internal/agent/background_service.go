package agent

import (
	"context"
	"fmt"

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
)

var logger bgsvc.Logger

type KmAgentService struct {
	Collector collector.KmCollector
	Configs   otelcol.CollectorSettings
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
	fmt.Println("Starting to do async work")
	go p.asyncWork()
	return nil
}

func (p *KmAgentService) Stop(s bgsvc.Service) error {
	fmt.Println("stopping the KloudMate Agent !")
	close(p.Exit)
	return nil
}
