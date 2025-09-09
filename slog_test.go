package logging

import (
	"log/slog"
	"testing"
)

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
