package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kardianos/service"
	bgsvc "github.com/kardianos/service"

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

	if len(os.Args) > 1 {
		verb := os.Args[1]

		switch verb {
		case "install":
			if err := s.Install(); err != nil {
				fmt.Println("failed to install:", err)
				return
			}
			fmt.Printf("[INFO] : service \"%s\" installed.\n", svcConfig.DisplayName)
			return

		case "uninstall":
			if err := s.Uninstall(); err != nil {
				fmt.Println("Failed to uninstall:", err)
				return
			}

			fmt.Printf("[INFO] : service \"%s\" uninstalled.\n", svcConfig.DisplayName)
			return

		case "start":
			if err := s.Start(); err != nil {
				fmt.Println("Failed to start:", err)
				return
			}

			fmt.Printf("[INFO] : service \"%s\" started.\n", svcConfig.DisplayName)
			return

		case "stop":
			if err := s.Stop(); err != nil {
				fmt.Println("Failed to stop:", err)
				return
			}

			fmt.Printf("[INFO] : service \"%s\" stopped.\n", svcConfig.DisplayName)
			return

		case "restart":
			if err := s.Restart(); err != nil {
				fmt.Println("Failed to restart:", err)
				return
			}

			fmt.Printf("[INFO] : service \"%s\" restarted.\n", svcConfig.DisplayName)
			return

		case "dry-run":
			if err := s.Run(); err != nil {
				fmt.Println("Failed to dry-run:", err)
				return
			}

			fmt.Printf("[INFO] : service \"%s\" dry-run started.\n", svcConfig.DisplayName)
			return

		default:
			fmt.Printf("Options for \"%s\": (install | uninstall | start | stop | dry-run)\n", os.Args[0])
			return
		}
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}
