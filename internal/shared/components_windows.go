//go:build windows

package shared

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/iisreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlserverreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowseventlogreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsperfcountersreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsservicereceiver"

	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/receiver"
)

// Components returns the collector factories for Windows.
func Components() (otelcol.Factories, error) {
	extensions := BaseExtensionFactories()
	receivers := BaseReceiverFactories()
	exporters := BaseExporterFactories()
	processors := BaseProcessorFactories()
	connectors := BaseConnectorFactories()

	receivers = append(receivers, windowsReceivers()...)

	return BuildFactories(extensions, receivers, exporters, processors, connectors)
}

// windowsReceivers returns Windows-specific receiver factories.
func windowsReceivers() []receiver.Factory {
	return []receiver.Factory{
		windowseventlogreceiver.NewFactory(),
		windowsperfcountersreceiver.NewFactory(),
		iisreceiver.NewFactory(),
		windowsservicereceiver.NewFactory(),
		sqlserverreceiver.NewFactory(),
	}
}
