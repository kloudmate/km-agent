//go:build kubernetes

package shared

import (
	ebpfreceiver "components.kloudmate.com/receiver/ebpfreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/cgroupruntimeextension"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8seventsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8slogreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sobjectsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver"

	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"
)

// Components returns the collector factories for Kubernetes builds.
func Components() (otelcol.Factories, error) {
	extensions := BaseExtensionFactories()
	exporters := BaseExporterFactories()
	processors := BaseProcessorFactories()
	connectors := BaseConnectorFactories()

	receivers := BaseReceiverFactories()
	receivers = append(receivers, kubernetesReceivers()...)

	extensions = append(extensions, kubernetesExtensions()...)

	processors = append(processors, kubernetesProcessors()...)

	return BuildFactories(extensions, receivers, exporters, processors, connectors)
}

// kubernetesExtensions returns Kubernetes-specific extension factories.
func kubernetesExtensions() []extension.Factory {
	return []extension.Factory{
		cgroupruntimeextension.NewFactory(),
	}
}

// kubernetesReceivers returns Kubernetes-specific receiver factories.
func kubernetesReceivers() []receiver.Factory {
	return []receiver.Factory{
		k8sclusterreceiver.NewFactory(),
		k8sobjectsreceiver.NewFactory(),
		kubeletstatsreceiver.NewFactory(),
		k8seventsreceiver.NewFactory(),
		k8slogreceiver.NewFactory(),
		ebpfreceiver.NewFactory(),
	}
}

// kubernetesProcessors returns Kubernetes-specific processor factories.
func kubernetesProcessors() []processor.Factory {
	return []processor.Factory{
		k8sattributesprocessor.NewFactory(),
	}
}
