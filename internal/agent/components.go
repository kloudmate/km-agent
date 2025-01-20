package agent

import (
	dockerstatsreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver"
	hostmetricsreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	forwardconnector "go.opentelemetry.io/collector/connector/forwardconnector"
	"go.opentelemetry.io/collector/exporter"
	debugexporter "go.opentelemetry.io/collector/exporter/debugexporter"
	nopexporter "go.opentelemetry.io/collector/exporter/nopexporter"
	otlpexporter "go.opentelemetry.io/collector/exporter/otlpexporter"
	otlphttpexporter "go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/extension"
	memorylimiterextension "go.opentelemetry.io/collector/extension/memorylimiterextension"
	zpagesextension "go.opentelemetry.io/collector/extension/zpagesextension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	batchprocessor "go.opentelemetry.io/collector/processor/batchprocessor"
	memorylimiterprocessor "go.opentelemetry.io/collector/processor/memorylimiterprocessor"
	"go.opentelemetry.io/collector/receiver"
	nopreceiver "go.opentelemetry.io/collector/receiver/nopreceiver"
	otlpreceiver "go.opentelemetry.io/collector/receiver/otlpreceiver"
)

// responsible for injecting all the extensions, receivers, processors, exporters to the collectors
func components() (otelcol.Factories, error) {
	var err error
	factories := otelcol.Factories{}

	factories.Extensions, err = extension.MakeFactoryMap(
		memorylimiterextension.NewFactory(),
		zpagesextension.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}
	factories.ExtensionModules = make(map[component.Type]string, len(factories.Extensions))
	factories.ExtensionModules[memorylimiterextension.NewFactory().Type()] = "go.opentelemetry.io/collector/extension/memorylimiterextension v0.114.0"
	factories.ExtensionModules[zpagesextension.NewFactory().Type()] = "go.opentelemetry.io/collector/extension/zpagesextension v0.114.0"

	factories.Receivers, err = receiver.MakeFactoryMap(
		nopreceiver.NewFactory(),
		otlpreceiver.NewFactory(),
		hostmetricsreceiver.NewFactory(),
		dockerstatsreceiver.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}
	factories.ReceiverModules = make(map[component.Type]string, len(factories.Receivers))
	factories.ReceiverModules[nopreceiver.NewFactory().Type()] = "go.opentelemetry.io/collector/receiver/nopreceiver v0.114.0"
	factories.ReceiverModules[otlpreceiver.NewFactory().Type()] = "go.opentelemetry.io/collector/receiver/otlpreceiver v0.114.0"
	factories.ReceiverModules[hostmetricsreceiver.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver v0.112.0"
	factories.ReceiverModules[dockerstatsreceiver.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver v0.112.0"

	factories.Exporters, err = exporter.MakeFactoryMap(
		debugexporter.NewFactory(),
		nopexporter.NewFactory(),
		otlpexporter.NewFactory(),
		otlphttpexporter.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}
	factories.ExporterModules = make(map[component.Type]string, len(factories.Exporters))
	factories.ExporterModules[debugexporter.NewFactory().Type()] = "go.opentelemetry.io/collector/exporter/debugexporter v0.114.0"
	factories.ExporterModules[nopexporter.NewFactory().Type()] = "go.opentelemetry.io/collector/exporter/nopexporter v0.114.0"
	factories.ExporterModules[otlpexporter.NewFactory().Type()] = "go.opentelemetry.io/collector/exporter/otlpexporter v0.114.0"
	factories.ExporterModules[otlphttpexporter.NewFactory().Type()] = "go.opentelemetry.io/collector/exporter/otlphttpexporter v0.114.0"

	factories.Processors, err = processor.MakeFactoryMap(
		batchprocessor.NewFactory(),
		memorylimiterprocessor.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}
	factories.ProcessorModules = make(map[component.Type]string, len(factories.Processors))
	factories.ProcessorModules[batchprocessor.NewFactory().Type()] = "go.opentelemetry.io/collector/processor/batchprocessor v0.114.0"
	factories.ProcessorModules[memorylimiterprocessor.NewFactory().Type()] = "go.opentelemetry.io/collector/processor/memorylimiterprocessor v0.114.0"

	factories.Connectors, err = connector.MakeFactoryMap(
		forwardconnector.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}
	factories.ConnectorModules = make(map[component.Type]string, len(factories.Connectors))
	factories.ConnectorModules[forwardconnector.NewFactory().Type()] = "go.opentelemetry.io/collector/connector/forwardconnector v0.114.0"

	return factories, nil
}
