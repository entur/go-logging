package logging

import (
	"log/slog"
	"strings"

	"github.com/rs/zerolog"
)

const disabledSlogLevel slog.Level = 999 // slog.Level(zerolog.Disabled)

func convertStrToZLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "fatal", "ftl":
		return FatalLevel
	case "panic", "pnc":
		return PanicLevel
	case "error", "err":
		return ErrorLevel
	case "warning", "wrn":
		return WarnLevel
	case "info", "inf":
		return InfoLevel
	case "debug", "dbg":
		return DebugLevel
	case "trace", "trc":
		fallthrough
	default:
		return TraceLevel
	}
}

func convertZLogLevelToSLog(level zerolog.Level) slog.Level {
	switch level {
	case FatalLevel, PanicLevel, ErrorLevel:
		return slog.LevelError
	case WarnLevel:
		return slog.LevelWarn
	case InfoLevel:
		return slog.LevelInfo
	case DebugLevel, TraceLevel, NoLevel:
		return slog.LevelDebug
	default:
		return disabledSlogLevel
	}
}

func convertSLogLevelToZLog(level slog.Level) zerolog.Level {
	switch level {
	case slog.LevelError:
		return ErrorLevel
	case slog.LevelWarn:
		return WarnLevel
	case slog.LevelInfo:
		return InfoLevel
	case slog.LevelDebug:
		return DebugLevel
	default:
		return DebugLevel
	}
}
