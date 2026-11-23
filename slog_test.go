package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"reflect"
	"strings"
	"testing"
	"time"
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

// decodeLine decodes the single JSON line written to buf.
func decodeLine(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not a JSON object: %v\n%s", err, buf.String())
	}

	return got
}

func newTestSlogger(opts ...Option) (*slog.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	opts = append([]Option{WithWriter(buf), WithNoTimestamp()}, opts...)

	return slog.New(NewSlogHandler(opts...)), buf
}

type testValuer struct{}

func (testValuer) LogValue() slog.Value {
	return slog.GroupValue(slog.String("resolved", "yes"))
}

func TestHandleRecordAttrs(t *testing.T) {
	slogger, buf := newTestSlogger()

	slogger.Warn("rejected", "cause", "expired", slog.Int("code", 7))

	want := map[string]any{"severity": "warning", "message": "rejected", "cause": "expired", "code": float64(7)}
	if got := decodeLine(t, buf); !reflect.DeepEqual(got, want) {
		t.Errorf("record attrs not logged\ngot: %v\nwant: %v", got, want)
	}
}

func TestSlogGroups(t *testing.T) {
	tests := []struct {
		title string
		log   func(l *slog.Logger)
		want  map[string]any
	}{
		{
			title: "record attrs nested under WithGroup",
			log:   func(l *slog.Logger) { l.WithGroup("g").Warn("m", "a", 1) },
			want:  map[string]any{"g": map[string]any{"a": float64(1)}},
		},
		{
			title: "WithAttrs after WithGroup nested",
			log:   func(l *slog.Logger) { l.WithGroup("g").With("a", 1).Warn("m") },
			want:  map[string]any{"g": map[string]any{"a": float64(1)}},
		},
		{
			title: "WithAttrs before WithGroup stays at top level",
			log:   func(l *slog.Logger) { l.With("a", 1).WithGroup("g").Warn("m", "b", 2) },
			want:  map[string]any{"a": float64(1), "g": map[string]any{"b": float64(2)}},
		},
		{
			title: "nested groups",
			log:   func(l *slog.Logger) { l.WithGroup("g").With("a", 1).WithGroup("h").Warn("m", "b", 2) },
			want: map[string]any{"g": map[string]any{
				"a": float64(1),
				"h": map[string]any{"b": float64(2)},
			}},
		},
		{
			title: "empty group name is inlined",
			log:   func(l *slog.Logger) { l.WithGroup("").Warn("m", "a", 1) },
			want:  map[string]any{"a": float64(1)},
		},
		{
			title: "group without attrs is omitted",
			log:   func(l *slog.Logger) { l.WithGroup("g").Warn("m") },
			want:  map[string]any{},
		},
		{
			title: "nested group without attrs is omitted",
			log:   func(l *slog.Logger) { l.WithGroup("g").With("a", 1).WithGroup("h").Warn("m") },
			want:  map[string]any{"g": map[string]any{"a": float64(1)}},
		},
		{
			title: "slog.Group value nests",
			log:   func(l *slog.Logger) { l.Warn("m", slog.Group("g", "a", 1)) },
			want:  map[string]any{"g": map[string]any{"a": float64(1)}},
		},
		{
			title: "slog.Group with empty key is inlined",
			log:   func(l *slog.Logger) { l.Warn("m", slog.Group("", "a", 1)) },
			want:  map[string]any{"a": float64(1)},
		},
		{
			title: "empty slog.Group is omitted",
			log:   func(l *slog.Logger) { l.Warn("m", slog.Group("g")) },
			want:  map[string]any{},
		},
		{
			title: "slog.Group containing only empty attrs is omitted",
			log:   func(l *slog.Logger) { l.Warn("m", slog.Group("g", slog.Attr{})) },
			want:  map[string]any{},
		},
		{
			title: "LogValuer is resolved",
			log:   func(l *slog.Logger) { l.Warn("m", "v", testValuer{}) },
			want:  map[string]any{"v": map[string]any{"resolved": "yes"}},
		},
		{
			title: "LogValuer is resolved in WithAttrs",
			log:   func(l *slog.Logger) { l.With("v", testValuer{}).Warn("m") },
			want:  map[string]any{"v": map[string]any{"resolved": "yes"}},
		},
		{
			title: "empty attr is ignored",
			log:   func(l *slog.Logger) { l.Warn("m", slog.Attr{}, "a", 1) },
			want:  map[string]any{"a": float64(1)},
		},
		{
			title: "empty attr in WithAttrs is ignored",
			log:   func(l *slog.Logger) { l.With(slog.Attr{}).Warn("m") },
			want:  map[string]any{},
		},
		{
			title: "derived handlers do not leak into the parent",
			log: func(l *slog.Logger) {
				_ = l.WithGroup("g").With("a", 1)
				l.Warn("m")
			},
			want: map[string]any{},
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			slogger, buf := newTestSlogger()

			test.log(slogger)

			got := decodeLine(t, buf)
			delete(got, "severity")
			delete(got, "message")
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("unexpected output\ngot: %v\nwant: %v", got, test.want)
			}
		})
	}
}

func TestSlogDerivedHandlerKeepsLevel(t *testing.T) {
	ctx := context.Background()
	handler := NewSlogHandler(WithLevel(ErrorLevel), WithNoTimestamp())

	derived := map[string]slog.Handler{
		"WithAttrs": handler.WithAttrs([]slog.Attr{slog.Int("a", 1)}),
		"WithGroup": handler.WithGroup("g"),
	}

	for title, h := range derived {
		if h.Enabled(ctx, slog.LevelWarn) {
			t.Errorf("%s: derived handler lost parent's level, warn is enabled", title)
		}
		if !h.Enabled(ctx, slog.LevelError) {
			t.Errorf("%s: derived handler has error disabled", title)
		}
	}
}

func TestSlogDerivedHandlerKeepsNoTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := NewSlogHandler(WithWriter(buf), WithNoTimestamp(), WithLevel(DebugLevel))

	for _, h := range []slog.Handler{
		handler.WithAttrs([]slog.Attr{slog.Int("a", 1)}),
		handler.WithGroup("g"),
		handler.WithGroup("g").WithAttrs([]slog.Attr{slog.Int("a", 1)}),
	} {
		buf.Reset()
		if err := h.Handle(context.Background(), slog.NewRecord(time.Now(), slog.LevelWarn, "m", 0)); err != nil {
			t.Fatal(err)
		}
		if got := decodeLine(t, buf); got["timestamp"] != nil {
			t.Errorf("derived handler wrote a timestamp: %v", got)
		}
	}
}

func TestSlogReservedTimestamp(t *testing.T) {
	at := time.Date(2024, 3, 4, 5, 6, 7, 0, time.UTC)
	const stamp = "2024-03-04T05:06:07Z"

	tests := []struct {
		title string
		log   func(h slog.Handler, r slog.Record) error
		want  map[string]any
	}{
		{
			title: "WithGroup record attrs",
			log: func(h slog.Handler, r slog.Record) error {
				r.AddAttrs(slog.Int("a", 1))

				return h.WithGroup("timestamp").Handle(context.Background(), r)
			},
		},
		{
			title: "WithGroup WithAttrs",
			log: func(h slog.Handler, r slog.Record) error {
				return h.WithGroup("timestamp").WithAttrs([]slog.Attr{slog.Int("a", 1)}).Handle(context.Background(), r)
			},
		},
		{
			title: "slog.Group record attrs",
			log: func(h slog.Handler, r slog.Record) error {
				r.AddAttrs(slog.Group("timestamp", "a", 1))

				return h.Handle(context.Background(), r)
			},
		},
		{
			title: "slog.Group WithAttrs",
			log: func(h slog.Handler, r slog.Record) error {
				return h.WithAttrs([]slog.Attr{slog.Group("timestamp", "a", 1)}).Handle(context.Background(), r)
			},
		},
		{
			title: "nested timestamp is kept",
			log: func(h slog.Handler, r slog.Record) error {
				r.AddAttrs(slog.String("timestamp", "v"))

				return h.WithGroup("g").WithAttrs([]slog.Attr{slog.String("timestamp", "w")}).Handle(context.Background(), r)
			},
			want: map[string]any{"g": map[string]any{"timestamp": "v"}},
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			buf := &bytes.Buffer{}
			h := NewSlogHandler(WithWriter(buf))

			if err := test.log(h, slog.NewRecord(at, slog.LevelWarn, "m", 0)); err != nil {
				t.Fatal(err)
			}

			if n := strings.Count(buf.String(), `"timestamp":`); n != 1 && test.want == nil {
				t.Errorf("want exactly one top-level timestamp key, got %d\n%s", n, buf.String())
			}

			got := decodeLine(t, buf)
			if got["timestamp"] != stamp && test.want == nil {
				t.Errorf("handler timestamp was replaced\ngot: %v\nwant: %s", got["timestamp"], stamp)
			}
			if test.want != nil && !reflect.DeepEqual(got["g"], test.want["g"]) {
				t.Errorf("nested timestamp lost\ngot: %v\nwant: %v", got["g"], test.want["g"])
			}
		})
	}
}
