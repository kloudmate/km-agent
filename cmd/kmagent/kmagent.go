package main

import (
	"fmt"
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
	errs := make(chan error, 56)
	logger, err = s.Logger(errs)
	if err != nil {
		log.Fatal(err)
	}
	prg.Svclogger = logger
	if err != nil {
		log.Fatal(err)
	}

	app := &cli.App{
		Name:     "kmagent",
		Usage:    "KloudMate Agent for auto instrumentation",
		Flags:    prg.CliArgs(),
		Commands: prg.CliCommands(s),
	}
	prg.InitCollector(app)
	// prg.ApplyAgentConfig(cli.NewContext(app, nil, nil))

	// show help when no command specified
	app.Action = func(c *cli.Context) error {
		return cli.ShowAppHelp(c)
	}

	if err := app.Run(os.Args); err != nil {
		// fmt.Println(string(prg.Collector.ErrBuff.Bytes()))
		fmt.Println("=====================")
		fmt.Println(err)
		// log.Fatal(err)
	}
	// defer fmt.Println(string(prg.Collector.ErrBuff.Bytes()))

}
