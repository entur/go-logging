package logging

import (
	"log/slog"
	"testing"

	"github.com/rs/zerolog"
)

func TestZtoS(t *testing.T) {
	if slog.LevelError != levelZlogToSlog(zerolog.FatalLevel) {
		t.Error("Fatal must be Error")
	}
	if slog.LevelWarn != levelZlogToSlog(zerolog.WarnLevel) {
		t.Error("Warn must be Warn")
	}
	if slog.LevelInfo != levelZlogToSlog(zerolog.InfoLevel) {
		t.Error("Info must be Info")
	}
}

func TestStoZ(t *testing.T) {
	if zerolog.ErrorLevel != levelSlogToZlog(slog.LevelError) {
		t.Error("Error must be Error")
	}
	if zerolog.WarnLevel != levelSlogToZlog(slog.LevelWarn) {
		t.Error("Error must be Error")
	}
	if zerolog.DebugLevel != levelSlogToZlog(slog.LevelDebug) {
		t.Error("Debug must be Debug")
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

	setLevel("fatal")
	if zerolog.GlobalLevel() != zerolog.FatalLevel {
		t.Error("fatal is Fatal")
	}
}
