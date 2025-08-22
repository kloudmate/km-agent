//go:build windows
// +build windows

package logger

import (
	"os"

	"github.com/charmbracelet/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sys/windows/svc/eventlog"
)

type KmLogger struct {
	Logger    *zap.Logger
	WinLogger *WindowsEventLogCore
}

// SetupLogger returns logger for windows systems with event logging support
func SetupLogger() *KmLogger {

	eventLogCore, err := NewWindowsEventLogCore("kmagent", zapcore.InfoLevel)
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
	k.Logger.Sync()
	k.WinLogger.Close()
}

// Windows Event Log Core for Zap
type WindowsEventLogCore struct {
	zapcore.LevelEnabler
	elog    *eventlog.Log
	encoder zapcore.Encoder
}

func NewWindowsEventLogCore(serviceName string, enabler zapcore.LevelEnabler) (*WindowsEventLogCore, error) {
	elog, err := eventlog.Open(serviceName)
	if err != nil {
		err = eventlog.InstallAsEventCreate(serviceName, eventlog.Error|eventlog.Warning|eventlog.Info)
		if err != nil {
			return nil, err
		}
		elog, err = eventlog.Open(serviceName)
		if err != nil {
			return nil, err
		}
	}

	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	return &WindowsEventLogCore{
		LevelEnabler: enabler,
		elog:         elog,
		encoder:      encoder,
	}, nil
}

func (w *WindowsEventLogCore) With(fields []zapcore.Field) zapcore.Core {
	return &WindowsEventLogCore{
		LevelEnabler: w.LevelEnabler,
		elog:         w.elog,
		encoder:      w.encoder.Clone(),
	}
}

func (w *WindowsEventLogCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if w.Enabled(entry.Level) {
		return checked.AddCore(entry, w)
	}
	return checked
}

func (w *WindowsEventLogCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// Encode the log entry
	buf, err := w.encoder.EncodeEntry(entry, fields)
	if err != nil {
		return err
	}

	message := buf.String()

	// Map Zap levels to Windows Event Log levels
	switch entry.Level {
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel, zapcore.ErrorLevel:
		return w.elog.Error(3, "kmagent: "+message)
	case zapcore.WarnLevel:
		return w.elog.Warning(2, "kmagent: "+message)
	default:
		return w.elog.Info(1, "kmagent: "+message)
	}
}

func (w *WindowsEventLogCore) Sync() error {
	return nil
}

func (w *WindowsEventLogCore) Close() error {
	w.elog.Close()
	return nil
}
