package shared

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/cumulativetodeltaprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/deltatorateprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/redactionprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/journaldreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8seventsreceiver"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage/filestorage"
	"go.opentelemetry.io/collector/extension/memorylimiterextension"
	"go.opentelemetry.io/collector/extension/zpagesextension"

	"go.opentelemetry.io/collector/exporter/debugexporter"
	"go.opentelemetry.io/collector/exporter/nopexporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"

	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/processor/memorylimiterprocessor"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachereceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscloudwatchmetricsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscloudwatchreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscontainerinsightreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awsecscontainermetricsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azuremonitorreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/elasticsearchreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/googlecloudmonitoringreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/httpcheckreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkametricsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver"
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
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlqueryreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlserverreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/syslogreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver"

	"go.opentelemetry.io/collector/connector"

	"go.opentelemetry.io/collector/exporter"

	"go.opentelemetry.io/collector/extension"

	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"
)

// responsible for injecting all the extensions, receivers, processors, exporters to the collectors
func Components() (otelcol.Factories, error) {
	var err error
	factories := otelcol.Factories{}

	factories.Extensions, err = otelcol.MakeFactoryMap[extension.Factory](
		memorylimiterextension.NewFactory(),
		zpagesextension.NewFactory(),
		filestorage.NewFactory(),
		healthcheckextension.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.Receivers, err = otelcol.MakeFactoryMap[receiver.Factory](
		otlpreceiver.NewFactory(),
		hostmetricsreceiver.NewFactory(),
		dockerstatsreceiver.NewFactory(),
		apachereceiver.NewFactory(),
		elasticsearchreceiver.NewFactory(),
		filelogreceiver.NewFactory(),
		k8sclusterreceiver.NewFactory(),
		fluentforwardreceiver.NewFactory(),
		kafkametricsreceiver.NewFactory(),
		mongodbatlasreceiver.NewFactory(),
		mysqlreceiver.NewFactory(),
		nginxreceiver.NewFactory(),
		oracledbreceiver.NewFactory(),
		postgresqlreceiver.NewFactory(),
		rabbitmqreceiver.NewFactory(),
		redisreceiver.NewFactory(),
		sqlserverreceiver.NewFactory(),
		prometheusreceiver.NewFactory(),
		mongodbreceiver.NewFactory(),
		kubeletstatsreceiver.NewFactory(),
		journaldreceiver.NewFactory(),
		k8seventsreceiver.NewFactory(),
		awscloudwatchmetricsreceiver.NewFactory(),
		awscloudwatchreceiver.NewFactory(),
		awscontainerinsightreceiver.NewFactory(),
		awsecscontainermetricsreceiver.NewFactory(),
		azuremonitorreceiver.NewFactory(),
		googlecloudmonitoringreceiver.NewFactory(),
		httpcheckreceiver.NewFactory(),
		netflowreceiver.NewFactory(),
		saphanareceiver.NewFactory(),
		sqlqueryreceiver.NewFactory(),
		syslogreceiver.NewFactory(),
		vcenterreceiver.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.Exporters, err = otelcol.MakeFactoryMap[exporter.Factory](
		debugexporter.NewFactory(),
		nopexporter.NewFactory(),
		otlpexporter.NewFactory(),
		otlphttpexporter.NewFactory(),
	)

	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.Processors, err = otelcol.MakeFactoryMap[processor.Factory](
		batchprocessor.NewFactory(),
		resourceprocessor.NewFactory(),
		resourcedetectionprocessor.NewFactory(),
		attributesprocessor.NewFactory(),
		redactionprocessor.NewFactory(),
		probabilisticsamplerprocessor.NewFactory(),
		k8sattributesprocessor.NewFactory(),
		cumulativetodeltaprocessor.NewFactory(),
		deltatorateprocessor.NewFactory(),
		filterprocessor.NewFactory(),
		metricstransformprocessor.NewFactory(),
		memorylimiterprocessor.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.Connectors, err = otelcol.MakeFactoryMap[connector.Factory]()
	if err != nil {
		return otelcol.Factories{}, err
	}

	return factories, nil
}
