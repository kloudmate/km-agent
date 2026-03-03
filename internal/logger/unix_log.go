//go:build !windows
// +build !windows

package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

type KmLogger struct {
	Logger *zap.Logger
}

// SetupLogger returns simple logger for unix systems
func SetupLogger() *KmLogger {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(ParseLogLevel())
	logger, err := cfg.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create logger: %v\n", err)
		os.Exit(1)
	}
	return &KmLogger{Logger: logger}
}

func (k *KmLogger) MustCleanup() {
	k.Logger.Sync()
}
