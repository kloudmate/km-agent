package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings" 

	"github.com/kardianos/service"
	bgsvc "github.com/kardianos/service"
	cli "github.com/urfave/cli/v2"

	"github.com/kloudmate/km-agent/internal/agent"
)

var logger bgsvc.Logger


var CONFIG_FILE_URI string

func init() {
	if strings.Contains(os.Getenv("OS"), "Windows") {
		CONFIG_FILE_URI = filepath.Join(os.Getenv("PROGRAMDATA"), "km-agent", "config.yaml") 
	} else {
		CONFIG_FILE_URI = "/etc/km-agent/config.yaml"
	}
}

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
					apiKey := os.Getenv("KM_API_KEY")

					if apiKey == "" {
						return fmt.Errorf("KM_API_KEY environment variable not set")
					}
					configContent := fmt.Sprintf(`# This config file will be used by the KM-Agent on its first initialization.
                    receivers:
                    hostmetrics:
                        collection_interval: 25s
                    scrapers:
                        cpu:
                        disk:
                        processes:
                        process:
                        memory:
                        network:
                        filesystem:
                        load:
                        paging:
                    otlp:
                    protocols:
                        grpc:
                        http:

                    exporters:
                    debug:
                    otlphttp:
                       endpoint: https://otel.kloudmate.com:4318
                    headers:
                       Authorization: "%s"
                                 `, apiKey)

					
					configDir := filepath.Dir(CONFIG_FILE_URI)
					if _, err := os.Stat(configDir); os.IsNotExist(err) {
						err := os.MkdirAll(configDir, 0755)
						if err != nil {
							return fmt.Errorf("failed to create config directory: %w", err)
						}
					}

					
					configFile, err := os.OpenFile(CONFIG_FILE_URI, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
					if err != nil {
						return fmt.Errorf("failed to create config file: %w", err)
					}

					
					_, err = configFile.WriteString(configContent)
					configFile.Close() 

					if err != nil {
						return fmt.Errorf("failed to write config: %w", err)
					}

				
					err = service.Control(s, "install")
					if err != nil{
						fmt.Println("Failed to install the service")
					} else{
						fmt.Println("Successfully installed the service")
					}
					return err
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

					return service.Control(s, "start")
				},
			},
			{
				Name:  "stop",
				Usage: "Stop the service",
				Action: func(c *cli.Context) error {
					return service.Control(s, "stop")
				},
			},
			{
                Name:  "run",
                Usage: "Run the service in the foreground",
                Action: func(c *cli.Context) error {
                   return s.Run()
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