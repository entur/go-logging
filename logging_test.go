package logging

import (
	"log/slog"
	"testing"

	"github.com/rs/zerolog"
)

func TestZtoS(t *testing.T) {
	levels := map[slog.Level]zerolog.Level{
		slog.LevelError: zerolog.ErrorLevel,
		slog.LevelWarn:  zerolog.WarnLevel,
		slog.LevelDebug: zerolog.DebugLevel,
		slog.LevelInfo:  zerolog.InfoLevel,
	}

	for slevel, zlevel := range levels {
		if level := levelZlogToSlog(zlevel); level != slevel {
			t.Errorf("failure mapping zerolog %s to slog\ngot: %s\nwant: %s", zlevel, level, slevel)
		}
	}
}

func TestStoZ(t *testing.T) {
	levels := map[zerolog.Level]slog.Level{
		zerolog.ErrorLevel: slog.LevelError,
		zerolog.WarnLevel:  slog.LevelWarn,
		zerolog.DebugLevel: slog.LevelDebug,
	}

	for zlevel, slevel := range levels {
		if level := levelSlogToZlog(slevel); level != zlevel {
			t.Errorf("failure mapping slog %s to zerolog\ngot: %s\nwant: %s", slevel, level, zlevel)
		}
	}
}

func TestWithGroup(t *testing.T) {
	h := NewSlogHandler()
	slogger := slog.New(h)

	slog2 := slogger.WithGroup("")
	if slog2 != slogger {
		t.Error("Empty group is ignored")
	}
	slog3 := slogger.WithGroup("MyGroup")
	if slog3 == slogger {
		t.Error("Group is new instance")
	}
}

func Test_setLevel(t *testing.T) {

	level := zerolog.GlobalLevel()

	levels := map[string]zerolog.Level{
		"fatal":   zerolog.FatalLevel,
		"panic":   zerolog.PanicLevel,
		"error":   zerolog.ErrorLevel,
		"warning": zerolog.WarnLevel,
		"info":    zerolog.InfoLevel,
		"debug":   zerolog.DebugLevel,
	}

	for levelString, actualLevel := range levels {
		setLevel(levelString)
		if zerolog.GlobalLevel() != actualLevel {
			t.Errorf("%s is %s", levelString, actualLevel.String())
		}
	}

	t.Cleanup(func() {
		zerolog.SetGlobalLevel(level)
	})
}
