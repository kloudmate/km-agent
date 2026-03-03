package logger

import (
	"os"
	"strings"

	"go.uber.org/zap/zapcore"
)

// ParseLogLevel reads the KM_LOG_LEVEL environment variable and returns the
// corresponding zapcore.Level. Supported values: debug, info, warn, error.
// @amitava82: here it defaults to InfoLevel if unset or unrecognized.
func ParseLogLevel() zapcore.Level {
	switch strings.ToLower(os.Getenv("KM_LOG_LEVEL")) {
	case "debug":
		return zapcore.DebugLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
