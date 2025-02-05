package agent

import (
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

	// Commands
	installCommand   = "install"
	uninstallCommand = "uninstall"
	startCommand     = "start"
	stopCommand      = "stop"
)

// cliArgs are flags that are available in windows flavoured agent.
func (p *KmAgentService) CliArgs() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        modeFlag,
			Aliases:     []string{"m"},
			Value:       hostMode,
			Usage:       "Km Agent mode",
			Destination: &p.Mode,
		},
		&cli.StringFlag{
			Name:        keyFlag,
			Aliases:     []string{"key"},
			Value:       "",
			EnvVars:     []string{"KM_API_KEY"},
			Usage:       "used for kloudmate otel authentication",
			Destination: &p.AgentCfg.Key,
		},
		&cli.StringFlag{
			Name:        endpointFlag,
			Aliases:     []string{"collector-endpoint"},
			Value:       "",
			EnvVars:     []string{"KM_COLLECTOR_ENDPOINT"},
			Usage:       "for kloudmate collector endpoint",
			Destination: &p.AgentCfg.Endpoint,
		},
		&cli.StringFlag{
			Name:        intervalFlag,
			Aliases:     []string{"config-check-interval"},
			Value:       "10s",
			EnvVars:     []string{"KM_CONFIG_CHECK_INTERVAL"},
			Usage:       "for kloudmate otel config retrieval",
			Destination: &p.AgentCfg.Interval,
		},
		&cli.StringFlag{
			Name:        debugLevelFlag,
			Aliases:     []string{"debuglevel"},
			Value:       "normal",
			Usage:       "for kloudmate otel debugging",
			Destination: &p.AgentCfg.debugLevel,
		},
	}
}

func (p *KmAgentService) CliCommands(s bgsvc.Service) []*cli.Command {
	return []*cli.Command{
		{
			Name:  installCommand,
			Usage: "Install the service",
			Action: func(c *cli.Context) error {
				p.ApplyAgentConfig()
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
				p.ApplyAgentConfig()
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
