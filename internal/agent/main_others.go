//go:build !windows

package agent

import (
	"os"
)

var CONFIG_FILE_URI = os.Getenv("HOME") + "/.kloudmate/agent-config.yaml"
