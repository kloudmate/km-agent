package collector

import (
	"context"

	bgsvc "github.com/kardianos/service"
	"go.opentelemetry.io/collector/otelcol"
)

type KmCollector struct {
	otelcol.Collector
	exit chan struct{}
	svc  bgsvc.Service
}

func NewKmCollector(set otelcol.CollectorSettings) (*KmCollector, error) {

	col, err := otelcol.NewCollector(set)
	if err != nil {
		return nil, err
	}

	return &KmCollector{
		Collector: *col,
		exit:      make(chan struct{}),
	}, nil
}

func (p KmCollector) Start(s bgsvc.Service) error {
	// Start should not block. Do the actual work async.
	go p.Run(context.Background())
	return nil
}

func (p KmCollector) Stop(s bgsvc.Service) error {
	// Stop the service
	close(p.exit)
	return nil
}
