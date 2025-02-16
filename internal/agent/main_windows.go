//go:build windows

package agent

import (
	"os"
)

var AGENT_CONFIG_FILE_URI = os.Getenv("USERPROFILE") + "\\.kloudmate\\agent-config.yaml"
var HOST_CONFIG_FILE_URI = os.Getenv("USERPROFILE") + "\\.kloudmate\\host-col-config.yaml"
var DOCKER_CONFIG_FILE_URI = os.Getenv("USERPROFILE") + "\\.kloudmate\\docker-col-config.yaml"
