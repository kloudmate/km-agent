//go:build windows

package agent

import (
	"os"
)

var CONFIG_FILE_URI = os.Getenv("USERPROFILE") + "\\.kloudmate\\agent-config.yaml"
