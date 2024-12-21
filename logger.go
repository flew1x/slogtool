package slogtool

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime/debug"
	"time"

	"github.com/lmittmann/tint"
)

// LogLevel represents the severity of the log message.
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Logger interface defines methods for structured logging.
type Logger interface {
	Info(description string, attributes ...any)
	Debug(description string, attributes ...any)
	Error(description string, err error, attributes ...any)
	Warn(description string, attributes ...any)
	WithOperation(operation string) Logger
	StringAttr(attribute string, value string) any
	AnyAttr(attribute string, value any) any
	With(attributes ...any) Logger
	LogAndReturnError(message string, err error, attributes ...any) error
}

// Slogger is an implementation of Logger interface using slog.
type Slogger struct {
	logger *slog.Logger
}

func (l *Slogger) LogAndReturnError(message string, err error, attributes ...any) error {
	l.Error(message, err, attributes...)

	return err
}

// InitLogger initializes a new logger instance based on the provided
// mode and output strings. It returns a pointer to the initialized Slogger.
//
// If output is "stdout" or "stderr", it configures the logger for the specified
// log level to the respective standard output.
//
// If output is a file path, it configures the logger to log to that file.
func InitLogger(mode LogLevel, output string) *Slogger {
	options := &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.Kitchen,
	}

	var handler slog.Handler

	switch output {
	case "stdout":
		handler = getHandlerForOutput(os.Stdout, mode, options)
	case "stderr":
		handler = getHandlerForOutput(os.Stderr, mode, options)
	default:
		file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			panic(fmt.Sprintf("error opening log file %s: %s", output, err))
		}

		handler = getHandlerForOutput(file, mode, options)
	}

	buildInfo, _ := debug.ReadBuildInfo()
	if buildInfo == nil {
		buildInfo = &debug.BuildInfo{GoVersion: "unknown"}
	}

	return &Slogger{
		logger: slog.New(handler).With(
			slog.Group("program_info",
				slog.String("go_version", buildInfo.GoVersion),
			),
		),
	}
}

// getHandlerForOutput returns an appropriate handler based on the log level and options.
func getHandlerForOutput(output io.Writer, mode LogLevel, options *tint.Options) slog.Handler {
	switch mode {
	case LevelDebug:
		return tint.NewHandler(output, options)
	case LevelInfo:
		return slog.NewJSONHandler(output, &slog.HandlerOptions{Level: slog.LevelInfo})
	case LevelWarn:
		return slog.NewJSONHandler(output, &slog.HandlerOptions{Level: slog.LevelWarn})
	case LevelError:
		return slog.NewJSONHandler(output, &slog.HandlerOptions{Level: slog.LevelError})
	default:
		return tint.NewHandler(output, options)
	}
}

// Debug logs a debug level message.
func (l *Slogger) Debug(description string, attributes ...any) {
	l.logger.Debug(description, attributes...)
}

// Info logs an info level message.
func (l *Slogger) Info(description string, attributes ...any) {
	l.logger.Info(description, attributes...)
}

// Error logs an error level message with the error details.
func (l *Slogger) Error(description string, err error, attributes ...any) {
	if err == nil {
		attrs := append(attributes, slog.String("error", "nil"))

		l.logger.Error(description, attrs...)

		return
	}

	attrs := append(attributes, slog.String("error", err.Error()))
	l.logger.Error(description, attrs...)
}

// Warn logs a warn level message.
func (l *Slogger) Warn(description string, attributes ...any) {
	l.logger.Warn(description, attributes...)
}

// WithOperation returns a new logger with the given operation name.
func (l *Slogger) WithOperation(operation string) Logger {
	newLogger := *l
	newLogger.logger = newLogger.logger.With(slog.String("operation", operation))

	return &newLogger
}

// With returns a new logger with the given attributes.
func (l *Slogger) With(attributes ...any) Logger {
	newLogger := *l
	newLogger.logger = newLogger.logger.With(attributes...)

	return &newLogger
}

// StringAttr returns a string attribute for logging.
func (l *Slogger) StringAttr(attribute string, value string) any {
	return slog.String(attribute, value)
}

// AnyAttr returns an any type attribute for logging.
func (l *Slogger) AnyAttr(attribute string, value any) any {
	return slog.Any(attribute, value)
}
