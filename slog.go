package logging

import (
	"context"
	"log/slog"
	"maps"
	"slices"

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

// mergeInto adds attrs to m following the log/slog handler rules: LogValuers are resolved,
// empty Attrs and empty groups are dropped, and groups with an empty key are inlined.
func mergeInto(m map[string]any, as []slog.Attr) {
	for _, attr := range as {
		v := attr.Value.Resolve()

		if v.Kind() == slog.KindGroup {
			group := v.Group()
			switch {
			case len(group) == 0:
			case attr.Key == "":
				mergeInto(m, group)
			default:
				sub := map[string]any{}
				if existing, ok := m[attr.Key].(map[string]any); ok {
					sub = maps.Clone(existing)
				}

				mergeInto(sub, group)
				if len(sub) > 0 {
					m[attr.Key] = sub
				}
			}

			continue
		}

		if attr.Key == "" && v.Kind() == slog.KindAny && v.Any() == nil {
			continue
		}

		m[attr.Key] = v.Any()
	}
}

// cloneAndMergeAttrs returns a copy of attributes with as nested under groups.
// No group is created unless it ends up holding at least one attribute.
func cloneAndMergeAttrs(attributes map[string]any, groups []string, as []slog.Attr) map[string]any {
	if len(groups) == 0 {
		m := maps.Clone(attributes)
		if m == nil {
			m = map[string]any{}
		}

		mergeInto(m, as)

		return m
	}

	existing, _ := attributes[groups[0]].(map[string]any) // nolint:revive
	inner := cloneAndMergeAttrs(existing, groups[1:], as)
	if len(inner) == 0 {
		return maps.Clone(attributes) // callers may modify the result
	}

	m := maps.Clone(attributes)
	if m == nil {
		m = map[string]any{}
	}
	m[groups[0]] = inner

	return m
}

// mergeAttrs is cloneAndMergeAttrs for the handler's current groups. The top-level timestamp key,
// whether an attribute or a group, is reserved for the handler's own timestamp field.
func (h *SLogHandler) mergeAttrs(as []slog.Attr) map[string]any {
	attributes := cloneAndMergeAttrs(h.attributes, h.groups, as)
	delete(attributes, zerolog.TimestampFieldName)

	return attributes
}

func (h *SLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *SLogHandler) WithAttrs(as []slog.Attr) slog.Handler {
	attributes := h.mergeAttrs(as)
	if len(attributes) == 0 {
		return h
	}

	derived := *h
	derived.attributes = attributes

	return &derived
}

func (h *SLogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	derived := *h
	derived.groups = append(slices.Clone(h.groups), name)

	return &derived
}

func (h *SLogHandler) Handle(ctx context.Context, record slog.Record) error {
	attributes := h.attributes
	if record.NumAttrs() > 0 {
		as := make([]slog.Attr, 0, record.NumAttrs())
		record.Attrs(func(a slog.Attr) bool {
			as = append(as, a)

			return true
		})
		attributes = h.mergeAttrs(as)
	}

	c := h.logger.WithLevel(convertSLogLevelToZLog(record.Level)).Ctx(ctx).CallerSkipFrame(defaultSkipFrameCount)
	if !h.noTimestamp && !record.Time.IsZero() {
		c.Time(zerolog.TimestampFieldName, record.Time)
	}
	c.Fields(attributes).Msg(record.Message)

	return nil
}
