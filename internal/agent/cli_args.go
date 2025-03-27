package agent

import (
	"github.com/kloudmate/km-agent/internal/collector"
	cli "github.com/urfave/cli/v2"

	bgsvc "github.com/kardianos/service"
)

const (

	// Flags
	keyFlag        = "key"
	modeFlag       = "mode"
	debugLevelFlag = "debuglevel"
	endpointFlag   = "collector-endpoint"
	intervalFlag   = "config-check-interval"

	// Mode types
	hostMode      = "host"
	containerMode = "docker"

	// deafault telemetry endpoint
	defaultKmEndpoint = ""

	// Commands
	installCommand   = "install"
	uninstallCommand = "uninstall"
	startCommand     = "start"
	stopCommand      = "stop"
)

// cliArgs are flags that are available in windows flavoured agent.
func (svc *KmAgentService) CliArgs() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        modeFlag,
			Aliases:     []string{"m"},
			Value:       hostMode,
			Usage:       "Kloudmate Agent mode",
			Category:    "Agent Configuration",
			Destination: &svc.Mode,
		},
		&cli.StringFlag{
			Name:        keyFlag,
			Aliases:     []string{"k"},
			EnvVars:     []string{"KM_API_KEY"},
			Usage:       "used for kloudmate otel authentication",
			Category:    "Agent Configuration",
			Destination: &svc.AgentCfg.Key,
		},
		&cli.StringFlag{
			Name:        endpointFlag,
			Aliases:     []string{"endpoint"},
			Value:       defaultKmEndpoint,
			EnvVars:     []string{"KM_COLLECTOR_ENDPOINT"},
			Usage:       "collector endpoint to send telemetry data to",
			Category:    "Agent Configuration",
			Destination: &svc.AgentCfg.Endpoint,
		},
		&cli.StringFlag{
			Name:        intervalFlag,
			Aliases:     []string{"interval"},
			Value:       "10s",
			EnvVars:     []string{"KM_CONFIG_CHECK_INTERVAL"},
			Usage:       "configuration retrieval interval",
			Category:    "Agent Configuration",
			Destination: &svc.AgentCfg.Interval,
		},
		&cli.StringFlag{
			Name:        debugLevelFlag,
			Aliases:     []string{"debug-level"},
			Value:       "normal",
			Usage:       "for kloudmate otel debugging",
			Category:    "For Debugging Purposes",
			Destination: &svc.AgentCfg.debugLevel,
		},
	}
}

func (p *KmAgentService) CliCommands(s bgsvc.Service) []*cli.Command {
	return []*cli.Command{
		{
			Name:  installCommand,
			Usage: "Install the service",
			Action: func(c *cli.Context) error {
				p.setupAgent()
				return bgsvc.Control(s, installCommand)
			},
		},
		{
			Name:  uninstallCommand,
			Usage: "Uninstall the service",
			Action: func(c *cli.Context) error {
				return bgsvc.Control(s, uninstallCommand)
			},
		},
		{
			Name:  startCommand,
			Usage: "Start the service",
			Action: func(c *cli.Context) error {
				p.ApplyAgentConfig(c)
				p.Collector, _ = collector.NewKmCollector(p.Configs)
				err := s.Run()
				if err != nil {
					logger.Error(err)
					return err
				}
				return nil
			},
		},
		{
			Name:  stopCommand,
			Usage: "Stop the service",
			Action: func(c *cli.Context) error {
				return bgsvc.Control(s, stopCommand)
			},
		},
	}
}
