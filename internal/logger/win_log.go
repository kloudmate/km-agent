//go:build windows
// +build windows

package logger

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/kloudmate/km-agent/internal/windows"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type KmLogger struct {
	Logger    *zap.Logger
	WinLogger *windows.WindowsEventLogCore
}

// SetupLogger returns logger for windows systems with event logging support
func SetupLogger() *KmLogger {

	eventLogCore, err := windows.NewWindowsEventLogCore("kmagent", zapcore.InfoLevel)
	if err != nil {
		log.Errorf("failed to create windows event log core: %v\n", err)
	}
	defer eventLogCore.Close()

	consoleCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(os.Stdout),
		zapcore.InfoLevel,
	)

	// Combine otel core logger with win event logger
	combinedCore := zapcore.NewTee(consoleCore, eventLogCore)
	logger := zap.New(combinedCore)
	zap.ReplaceGlobals(logger)
	return &KmLogger{
		Logger:    logger,
		WinLogger: eventLogCore,
	}
}

func (k *KmLogger) MustCleanup() {
	k.WinLogger.Close()
}
