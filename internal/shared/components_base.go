package shared

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/connector/spanmetricsconnector"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage/filestorage"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/cumulativetodeltaprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/deltatorateprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbyattrsprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbytraceprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/isolationforestprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/redactionprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachereceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscloudwatchmetricsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscloudwatchreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscontainerinsightreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awsecscontainermetricsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azuremonitorreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/elasticsearchreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/googlecloudmonitoringreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/httpcheckreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkametricsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbatlasreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/netflowreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/nginxreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/oracledbreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/postgresqlreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/rabbitmqreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/saphanareceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/snmpreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlqueryreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/syslogreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver"

	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/debugexporter"
	"go.opentelemetry.io/collector/exporter/nopexporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/extension/memorylimiterextension"
	"go.opentelemetry.io/collector/extension/zpagesextension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/processor/memorylimiterprocessor"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/service/telemetry/otelconftelemetry"
)

// BaseExtensionFactories returns extension factories common to all platforms.
func BaseExtensionFactories() []extension.Factory {
	return []extension.Factory{
		memorylimiterextension.NewFactory(),
		zpagesextension.NewFactory(),
		filestorage.NewFactory(),
		healthcheckextension.NewFactory(),
	}
}

// BaseReceiverFactories returns receiver factories common to all platforms.
func BaseReceiverFactories() []receiver.Factory {
	return []receiver.Factory{
		// Core receivers
		otlpreceiver.NewFactory(),
		hostmetricsreceiver.NewFactory(),
		filelogreceiver.NewFactory(),
		prometheusreceiver.NewFactory(),
		httpcheckreceiver.NewFactory(),
		syslogreceiver.NewFactory(),
		netflowreceiver.NewFactory(),
		snmpreceiver.NewFactory(),
		fluentforwardreceiver.NewFactory(),
		
		apachereceiver.NewFactory(),
		elasticsearchreceiver.NewFactory(),
		kafkametricsreceiver.NewFactory(),
		mongodbatlasreceiver.NewFactory(),
		mongodbreceiver.NewFactory(),
		mysqlreceiver.NewFactory(),
		nginxreceiver.NewFactory(),
		oracledbreceiver.NewFactory(),
		postgresqlreceiver.NewFactory(),
		rabbitmqreceiver.NewFactory(),
		redisreceiver.NewFactory(),
		saphanareceiver.NewFactory(),
		sqlqueryreceiver.NewFactory(),

		vcenterreceiver.NewFactory(),
		awscloudwatchmetricsreceiver.NewFactory(),
		awscloudwatchreceiver.NewFactory(),
		awscontainerinsightreceiver.NewFactory(),
		awsecscontainermetricsreceiver.NewFactory(),
		azuremonitorreceiver.NewFactory(),
		googlecloudmonitoringreceiver.NewFactory(),
	}
}

// BaseExporterFactories returns exporter factories common to all platforms.
func BaseExporterFactories() []exporter.Factory {
	return []exporter.Factory{
		debugexporter.NewFactory(),
		nopexporter.NewFactory(),
		otlpexporter.NewFactory(),
		otlphttpexporter.NewFactory(),
	}
}

// BaseProcessorFactories returns processor factories common to all platforms.
func BaseProcessorFactories() []processor.Factory {
	return []processor.Factory{
		batchprocessor.NewFactory(),
		resourceprocessor.NewFactory(),
		resourcedetectionprocessor.NewFactory(),
		attributesprocessor.NewFactory(),
		redactionprocessor.NewFactory(),
		probabilisticsamplerprocessor.NewFactory(),
		cumulativetodeltaprocessor.NewFactory(),
		deltatorateprocessor.NewFactory(),
		filterprocessor.NewFactory(),
		metricstransformprocessor.NewFactory(),
		memorylimiterprocessor.NewFactory(),
		transformprocessor.NewFactory(),
		groupbyattrsprocessor.NewFactory(),
		spanprocessor.NewFactory(),
		groupbytraceprocessor.NewFactory(),
		isolationforestprocessor.NewFactory(),
	}
}

// BaseConnectorFactories returns connector factories common to all platforms.
func BaseConnectorFactories() []connector.Factory {
	return []connector.Factory{
		spanmetricsconnector.NewFactory(),
	}
}

// BuildFactories constructs otelcol.Factories from the provided component slices.
func BuildFactories(
	extensions []extension.Factory,
	receivers []receiver.Factory,
	exporters []exporter.Factory,
	processors []processor.Factory,
	connectors []connector.Factory,
) (otelcol.Factories, error) {
	var err error
	factories := otelcol.Factories{}

	factories.Extensions, err = otelcol.MakeFactoryMap[extension.Factory](extensions...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.Receivers, err = otelcol.MakeFactoryMap[receiver.Factory](receivers...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.Exporters, err = otelcol.MakeFactoryMap[exporter.Factory](exporters...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.Processors, err = otelcol.MakeFactoryMap[processor.Factory](processors...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.Connectors, err = otelcol.MakeFactoryMap[connector.Factory](connectors...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.Telemetry = otelconftelemetry.NewFactory()
	return factories, nil
}
