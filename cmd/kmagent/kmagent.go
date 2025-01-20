package main

import (
	"log"
	"os"

	"github.com/kardianos/service"
	bgsvc "github.com/kardianos/service"
	cli "github.com/urfave/cli/v2"

	"github.com/kloudmate/km-agent/internal/agent"
)

var logger bgsvc.Logger

func main() {
	var svcConfig = &bgsvc.Config{
		Name:        "kmagent",
		DisplayName: "KloudMate Agent",
		Description: "KloudMate Agent for auto instrumentation",
	}
	prg, err := agent.NewKmAgentService()
	if err != nil {
		log.Fatal(err)
	}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	app := &cli.App{
		Name:  "kmagent",
		Usage: "KloudMate Agent for auto instrumentation",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "mode",
				Aliases:     []string{"m"},
				Value:       "host",
				Usage:       "Km Agent mode",
				Destination: &prg.Mode,
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "install",
				Usage: "Install the service",
				Action: func(c *cli.Context) error {
					return service.Control(s, "install")
				},
			},
			{
				Name:  "uninstall",
				Usage: "Uninstall the service",
				Action: func(c *cli.Context) error {
					return service.Control(s, "uninstall")
				},
			},
			{
				Name:  "start",
				Usage: "Start the service",
				Action: func(c *cli.Context) error {

					err = s.Run()
					if err != nil {
						logger.Error(err)
						return err
					}
					return nil
				},
			},
			{
				Name:  "stop",
				Usage: "Stop the service",
				Action: func(c *cli.Context) error {
					return service.Control(s, "stop")
				},
			},
		},
	}

	app.Action = func(c *cli.Context) error {
		return cli.ShowAppHelp(c)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

	// err = s.Run()
	// if err != nil {
	// 	logger.Error(err)
	// }
}
