//go:build windows
// +build windows

package logger

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sys/windows/svc/eventlog"
)

const sourceName = "kmagent"

type KmLogger struct {
	Logger *zap.Logger
	syncer zapcore.WriteSyncer
}

// winlogWriter is a custom WriteSyncer that writes to the Windows Event Log.
type winlogWriter struct {
	log *eventlog.Log
}

// Write implements the io.Writer interface.
func (w *winlogWriter) Write(p []byte) (n int, err error) {
	// A simple way to determine the log level from the message.
	// A more robust implementation would parse the JSON log message.
	msg := string(p)
	if strings.Contains(msg, "\"level\":\"error\"") || strings.Contains(msg, "ERROR") {
		w.log.Error(1, msg)
	} else if strings.Contains(msg, "\"level\":\"warn\"") || strings.Contains(msg, "WARN") {
		w.log.Warning(1, msg)
	} else {
		w.log.Info(1, msg)
	}
	return len(p), nil
}

// Sync implements the zapcore.WriteSyncer interface.
func (w *winlogWriter) Sync() error {
	// The event log does not require explicit flushing.
	return nil
}

// NewWinlogWriter creates a new WinlogWriter.
func NewWinlogWriter(sourceName string) (zapcore.WriteSyncer, error) {
	log, err := eventlog.Open(sourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open event log: %w", err)
	}
	return &winlogWriter{log: log}, nil
}

// installEventSource installs the application as a Windows Event Log source.
func installEventSource(sourceName string) error {
	return eventlog.InstallAsEventCreate(sourceName, eventlog.Info|eventlog.Warning|eventlog.Error)
}

// SetupLogger returns logger for windows systems with event logging support
func SetupLogger() *KmLogger {

	if err := installEventSource(sourceName); err != nil {
		fmt.Printf("Failed to install event source: %v. Make sure you run with administrator privileges.\n", err)
	}

	winlogSyncer, err := NewWinlogWriter(sourceName)
	if err != nil {
		fmt.Printf("Failed to set up event log, falling back to the console logger: %v\n", err)
		winlogSyncer = zapcore.AddSync(os.Stdout)
	}

	loggerCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		winlogSyncer,
		zapcore.InfoLevel,
	)

	logger := zap.New(loggerCore)
	return &KmLogger{
		Logger: logger,
		syncer: winlogSyncer,
	}
}

func (k *KmLogger) MustCleanup() {
	k.Logger.Sync()
	k.syncer.(*winlogWriter).log.Close()
}
