package logging

import (
	"context"
	"log/slog"
	"maps"

	"github.com/rs/zerolog"
)

const defaultSkipFrameCount int = 3

type SLogHandler struct {
	logger      *zerolog.Logger
	level       slog.Level
	noTimestamp bool
	groups      []string
	attributes  map[string]any
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
					m2, ok = v2.(map[string]any)
				}

				if ok {
					m[attr.Key] = cloneAndMergeAttrs(m2, group)
				}
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
	delete(attributes, zerolog.TimestampFieldName)

	return &SLogHandler{
		logger:      h.logger,
		noTimestamp: h.noTimestamp,
		groups:      h.groups,
		attributes:  attributes,
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
		attributes:  h.attributes,
	}
}

func (h *SLogHandler) Handle(ctx context.Context, record slog.Record) error {
	c := h.logger.WithLevel(convertSLogLevelToZLog(record.Level)).Ctx(ctx).CallerSkipFrame(defaultSkipFrameCount)
	if !h.noTimestamp {
		c.Time(zerolog.TimestampFieldName, record.Time)
	}
	c.Fields(h.attributes).Msg(record.Message)

	return nil
}
