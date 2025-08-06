package agent

import (
	"fmt"

	"github.com/kloudmate/km-agent/internal/config"
	"github.com/kloudmate/km-agent/internal/shared"
	"go.opentelemetry.io/collector/otelcol"
)

func NewCollector(c *config.Config) (*otelcol.Collector, error) {
	collectorSettings := shared.CollectorInfoFactory(c.OtelConfigPath)
	fmt.Println("New collector created")

	return otelcol.NewCollector(collectorSettings)
}
