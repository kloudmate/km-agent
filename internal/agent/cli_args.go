package agent

import (
	cli "github.com/urfave/cli/v2"

	bgsvc "github.com/kardianos/service"
)

const (

	// Flags
	modeFlag = "mode"
	keyFlag  = "key"

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
			Aliases:     []string{"k"},
			Value:       "",
			Usage:       "used for kloudmate otel authentication",
			Destination: &p.Token,
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
