package shared

import (
	"fmt"
	"runtime/debug"
)

// getCollectorVersion is used to give underlying collector's version
func GetCollectorVersion() (specificDepVersion string) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Println("Could not read build info.")
		specificDepVersion = "unknown"
		return specificDepVersion
	}
	specificDepVersion = getSpecificDependencyVersion(info, "go.opentelemetry.io/collector")
	return specificDepVersion
}

func getSpecificDependencyVersion(info *debug.BuildInfo, modulePath string) string {
	for _, dep := range info.Deps {
		if dep.Path == modulePath {
			return dep.Version
		}
	}
	return ""
}
