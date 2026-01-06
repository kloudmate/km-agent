//go:build !windows && !kubernetes

package shared

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/cgroupruntimeextension"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/journaldreceiver"

	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/receiver"
)

// Components returns the collector factories for Linux host/Docker builds.
func Components() (otelcol.Factories, error) {
	extensions := BaseExtensionFactories()
	receivers := BaseReceiverFactories()
	exporters := BaseExporterFactories()
	processors := BaseProcessorFactories()
	connectors := BaseConnectorFactories()

	extensions = append(extensions, linuxExtensions()...)
	receivers = append(receivers, linuxReceivers()...)

	return BuildFactories(extensions, receivers, exporters, processors, connectors)
}

// linuxExtensions returns Linux-specific extension factories.
func linuxExtensions() []extension.Factory {
	return []extension.Factory{
		cgroupruntimeextension.NewFactory(),
	}
}

// linuxReceivers returns Linux-specific receiver factories.
func linuxReceivers() []receiver.Factory {
	return []receiver.Factory{
		dockerstatsreceiver.NewFactory(),
		journaldreceiver.NewFactory(),
	}
}
