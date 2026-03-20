// Package logger provides structured logging functionality using zerolog.
package logger

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Log is the global logger instance.
var Log zerolog.Logger

// Init initializes the global logger with the specified log level.
func Init(level string) {
	lvl := parseLevel(level)
	env := os.Getenv("API_ENV")

	// Formato del caller: [archivo:linea].
	zerolog.CallerMarshalFunc = func(_ uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + itoa(line)
	}

	if env == "production" {
		Log = zerolog.New(os.Stdout).Level(lvl).With().Timestamp().Caller().Logger()
	} else {
		Log = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).Level(lvl).With().Timestamp().Caller().Logger()
	}
}

func itoa(i int) string {
	if i < 10 {
		// #nosec G115 -- i is guaranteed to be < 10, safe for byte conversion
		return string([]byte{'0' + byte(i)})
	}
	// #nosec G115 -- i%10 is guaranteed to be < 10, safe for byte conversion
	return itoa(i/10) + string([]byte{'0' + byte(i%10)})
}

func parseLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// Debug returns a debug level log event.
func Debug() *zerolog.Event { return Log.Debug() }

// Info returns an info level log event.
func Info() *zerolog.Event { return Log.Info() }

// Warn returns a warning level log event.
func Warn() *zerolog.Event { return Log.Warn() }

// Error returns an error level log event.
func Error() *zerolog.Event { return Log.Error() }

// Fatal returns a fatal level log event.
func Fatal() *zerolog.Event { return Log.Fatal() }
