package logging

import (
	"context"
	"log/slog"
	"maps"

	"github.com/rs/zerolog"
)

func levelZlogToSlog(level zerolog.Level) slog.Level {
	switch level {
	case FatalLevel, PanicLevel, ErrorLevel:
		return slog.LevelError
	case WarnLevel:
		return slog.LevelWarn
	case InfoLevel:
		return slog.LevelInfo
	case DebugLevel, TraceLevel, NoLevel:
		return slog.LevelDebug
	}
	// Disabled
	return 999
}

func levelSlogToZlog(level slog.Level) zerolog.Level {
	switch level {
	case slog.LevelError:
		return ErrorLevel
	case slog.LevelWarn:
		return WarnLevel
	case slog.LevelInfo:
		return InfoLevel
	}
	return DebugLevel
}

type SLogHandler struct {
	logger      *zerolog.Logger
	level       slog.Level
	noTimestamp bool
	groups      []string
	attributes       map[string]any
}

func cloneAndMergeAttrs(attributes map[string]any, as []slog.Attr) map[string]any {
	m := maps.Clone(attributes)
	if m == nil {
		m = map[string]any{}
	}

	for _, attr := range as {
		v := attr.Value

		if v.Kind() == slog.KindGroup {
			group := v.Group()
			if len(group) > 0 {
				var m2 map[string]any

				v2, ok := m[attr.Key]
				if ok {
					m2, _ = v2.(map[string]any)
				}

				m[attr.Key] = cloneAndMergeAttrs(m2, group)
			}
		} else {
			m[attr.Key] = v.Any()
		}
	}

	return m
}

func (h *SLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *SLogHandler) WithAttrs(as []slog.Attr) slog.Handler {
	attributes := cloneAndMergeAttrs(h.attributes, as)
	if len(attributes) == 0 {
		return h
	}
	delete(attributes, "timestamp")

	return &SLogHandler{
		logger:      h.logger,
		noTimestamp: h.noTimestamp,
		groups:      h.groups,
		attributes:       attributes,
	}
}

func (h *SLogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	return &SLogHandler{
		logger:      h.logger,
		noTimestamp: h.noTimestamp,
		groups:      append(h.groups, name),
		attributes:       h.attributes,
	}
}

func (h *SLogHandler) Handle(ctx context.Context, record slog.Record) error {
	c := h.logger.WithLevel(levelSlogToZlog(record.Level)).Ctx(ctx).CallerSkipFrame(3)
	if !h.noTimestamp {
		c.Time(zerolog.TimestampFieldName, record.Time)
	}
	c.Fields(h.attributes).Msg(record.Message)

	return nil
}