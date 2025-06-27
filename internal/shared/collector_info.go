package shared

import (
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/otelcol"
)

func CollectorInfoFactory(cfgPath string) otelcol.CollectorSettings {

	info := component.BuildInfo{
		Command:     "kmagent",
		Description: "KloudMate Agent for OpenTelemetry",
		Version:     "1.0.0",
	}

	fmt.Println("config file ", cfgPath)

	return otelcol.CollectorSettings{
		BuildInfo:               info,
		Factories:               Components,
		DisableGracefulShutdown: true,
		ConfigProviderSettings: otelcol.ConfigProviderSettings{
			ResolverSettings: confmap.ResolverSettings{
				DefaultScheme: "env",
				URIs:          []string{cfgPath},
				ProviderFactories: []confmap.ProviderFactory{
					envprovider.NewFactory(),
					fileprovider.NewFactory(),
					yamlprovider.NewFactory(),
				},
			},
		},
		SkipSettingGRPCLogger: true, // Prevents gRPC from setting its own logger, uses zap instead
	}
}
