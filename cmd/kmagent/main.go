package main

import (
	"log"

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
	installErr := s.Install()
	if installErr == nil {
		logger.Infof("installing kmagent as a service...%s", installErr)
	}
	if installErr != nil {
		logger.Infof("kmagent already installed as a service! %s", installErr)
	}
	go func() {
		err = s.Start()
		if err != nil {
			logger.Errorf("failed to run the service : %s", err)
		}
	}()
	err = s.Run()
	if err != nil {
		logger.Errorf("failed to run the service : %s", err)
	}

}
