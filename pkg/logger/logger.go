package logger

import (
	"log/slog"
	"os"
)

const (
	levelDebug string = "debug"
	levelInfo  string = "info"
	levelWarn  string = "warn"
	levelError string = "error"
)

func NewLogger(loggerLevel string) *slog.Logger {
	switch loggerLevel {
	case levelDebug:
		return slog.New(
			slog.NewTextHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelDebug},
			),
		)
	case levelWarn:
		return slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelWarn},
			),
		)
	case levelError:
		return slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelError},
			),
		)
	case levelInfo:
		fallthrough
	default:
		return slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelInfo},
			),
		)
	}
}

func Error(err error) slog.Attr {
	return slog.Attr{Key: "error", Value: slog.StringValue(err.Error())}
}

func String(key, value string) slog.Attr {
	return slog.Attr{Key: key, Value: slog.StringValue(value)}
}

func Int64(key string, value int64) slog.Attr {
	return slog.Attr{Key: key, Value: slog.Int64Value(value)}
}

func Any(key string, value any) slog.Attr {
	return slog.Attr{Key: key, Value: slog.AnyValue(value)}
}
