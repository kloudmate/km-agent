//go:build !windows && !linux && !kubernetes

package shared

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/cgroupruntimeextension"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/iisreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/journaldreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8seventsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8slogreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sobjectsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlserverreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowseventlogreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsperfcountersreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsservicereceiver"

	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"
)

// Components returns the collector factories with ALL components.
// This is the fallback used when no platform-specific build tag is provided.
// It includes all components for maximum compatibility.
func Components() (otelcol.Factories, error) {
	extensions := BaseExtensionFactories()
	receivers := BaseReceiverFactories()
	exporters := BaseExporterFactories()
	processors := BaseProcessorFactories()
	connectors := BaseConnectorFactories()

	// Add all platform-specific components for fallback compatibility
	extensions = append(extensions, allPlatformExtensions()...)
	receivers = append(receivers, allPlatformReceivers()...)
	processors = append(processors, allPlatformProcessors()...)

	return BuildFactories(extensions, receivers, exporters, processors, connectors)
}

// allPlatformExtensions returns extensions from all platforms.
func allPlatformExtensions() []extension.Factory {
	return []extension.Factory{
		cgroupruntimeextension.NewFactory(),
	}
}

// allPlatformReceivers returns receivers from all platforms.
func allPlatformReceivers() []receiver.Factory {
	return []receiver.Factory{
		// Linux-specific
		dockerstatsreceiver.NewFactory(),
		journaldreceiver.NewFactory(),
		// Windows-specific
		windowseventlogreceiver.NewFactory(),
		windowsperfcountersreceiver.NewFactory(),
		iisreceiver.NewFactory(),
		windowsservicereceiver.NewFactory(),
		sqlserverreceiver.NewFactory(),
		// Kubernetes-specific
		k8sclusterreceiver.NewFactory(),
		k8sobjectsreceiver.NewFactory(),
		kubeletstatsreceiver.NewFactory(),
		k8seventsreceiver.NewFactory(),
		k8slogreceiver.NewFactory(),
	}
}

// allPlatformProcessors returns processors from all platforms.
func allPlatformProcessors() []processor.Factory {
	return []processor.Factory{
		k8sattributesprocessor.NewFactory(),
	}
}
