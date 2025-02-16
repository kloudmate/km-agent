package main

import (
	"log"
	"os"

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
		Arguments:   []string{"start"},
	}
	prg, err := agent.NewKmAgentService()
	if err != nil {
		log.Fatal(err)
	}
	s, err := bgsvc.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	app := &cli.App{
		Name:     "kmagent",
		Usage:    "KloudMate Agent for auto instrumentation",
		Flags:    prg.CliArgs(),
		Commands: prg.CliCommands(s),
	}

	// show help when no command specified
	app.Action = func(c *cli.Context) error {
		return cli.ShowAppHelp(c)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
