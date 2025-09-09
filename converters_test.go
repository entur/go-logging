package logging

import (
	"log/slog"
	"testing"

	"github.com/rs/zerolog"
)

func TestConvertZLogLevelToSLog(t *testing.T) {
	levels := map[slog.Level]zerolog.Level{
		slog.LevelError:   zerolog.ErrorLevel,
		slog.LevelWarn:    zerolog.WarnLevel,
		slog.LevelDebug:   zerolog.DebugLevel,
		slog.LevelInfo:    zerolog.InfoLevel,
		disabledSlogLevel: zerolog.Disabled,
	}

	for slevel, zlevel := range levels {
		if level := convertZLogLevelToSLog(zlevel); level != slevel {
			t.Errorf("failure mapping zerolog %s to slog\ngot: %s\nwant: %s", zlevel, level, slevel)
		}
	}
}

func TestConvertSLogLevelToZLog(t *testing.T) {
	levels := map[zerolog.Level]slog.Level{
		zerolog.ErrorLevel: slog.LevelError,
		zerolog.WarnLevel:  slog.LevelWarn,
		zerolog.InfoLevel:  slog.LevelInfo,
		zerolog.DebugLevel: slog.LevelDebug,
		zerolog.Disabled:   disabledSlogLevel,
	}

	for zlevel, slevel := range levels {
		if level := convertSLogLevelToZLog(slevel); level != zlevel {
			t.Errorf("failure mapping slog %s to zerolog\ngot: %s\nwant: %s", slevel, level, zlevel)
		}
	}
}

func TestConvertStrToZLogLevel(t *testing.T) {
	levels := map[string]zerolog.Level{
		"fatal":   zerolog.FatalLevel,
		"panic":   zerolog.PanicLevel,
		"error":   zerolog.ErrorLevel,
		"warning": zerolog.WarnLevel,
		"info":    zerolog.InfoLevel,
		"debug":   zerolog.DebugLevel,
		"trace":   zerolog.TraceLevel,
	}

	for strlevel, zlevel := range levels {
		if level := convertStrToZLogLevel(strlevel); level != zlevel {
			t.Errorf("failure mapping string level %s to zerolog\ngot: %s\nwant: %s", strlevel, level, zlevel)
		}
	}
}
