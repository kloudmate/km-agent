package agent

import (
	bgsvc "github.com/kardianos/service"
	"github.com/kloudmate/km-agent/internal/collector"
	"go.opentelemetry.io/collector/otelcol"
)

var logger bgsvc.Logger

// svcConfig comprised of the background service metadata
var svcConfig = &bgsvc.Config{
	Name:        "kmagent",
	DisplayName: "KloudMate Agent",
	Description: "OpenTelemetry auto instrumentation",
}

func NewAgentService(set otelcol.CollectorSettings) (bgsvc.Service, *collector.KmCollector, error) {
	col, err := collector.NewKmCollector(set)
	if err != nil {
		return nil, nil, err
	}
	svc, err := bgsvc.New(col, svcConfig)
	if err != nil {
		return nil, nil, err
	}
	return svc, col, nil
}
