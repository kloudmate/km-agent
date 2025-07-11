package agent

import (
	"fmt"
	"github.com/kloudmate/km-agent/internal/config"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/otelcol"
)

func NewCollector(c *config.Config) (*otelcol.Collector, error) {
	info := component.BuildInfo{
		Command:     "kmagent",
		Description: "KloudMate Agent for OpenTelemetry",
		Version:     "1.0.0",
	}

	fmt.Println("config file ", c.OtelConfigPath)

	set := otelcol.CollectorSettings{
		BuildInfo:               info,
		Factories:               components,
		DisableGracefulShutdown: true,
		ConfigProviderSettings: otelcol.ConfigProviderSettings{
			ResolverSettings: confmap.ResolverSettings{
				DefaultScheme: "env",
				URIs:          []string{c.OtelConfigPath},
				ProviderFactories: []confmap.ProviderFactory{
					envprovider.NewFactory(),
					fileprovider.NewFactory(),
					yamlprovider.NewFactory(),
				},
			},
		},
	}

	fmt.Println("New collector created")

	return otelcol.NewCollector(set)
}
