package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

const (
	testTraceHex = "4bf92f3577b34da6a3ce929d0e0e4736"
	testSpanHex  = "00f067aa0ba902b7"
)

func newSpanContext(t *testing.T, traceHex, spanHex string, flags trace.TraceFlags) context.Context {
	t.Helper()

	traceID, err := trace.TraceIDFromHex(traceHex)
	if err != nil {
		t.Fatalf("bad trace id: %v", err)
	}
	spanID, err := trace.SpanIDFromHex(spanHex)
	if err != nil {
		t.Fatalf("bad span id: %v", err)
	}

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: flags,
	})

	return trace.ContextWithSpanContext(context.Background(), sc)
}

func logLine(ctx context.Context, t *testing.T, opts ...Option) map[string]any {
	t.Helper()

	var buf bytes.Buffer
	opts = append([]Option{WithWriter(&buf), WithLevel(InfoLevel)}, opts...)
	logger := New(opts...)
	logger.Info().Ctx(ctx).Msg("hello")

	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal %q: %v", buf.String(), err)
	}

	return out
}

func TestTraceHookSampled(t *testing.T) {
	ctx := newSpanContext(t, testTraceHex, testSpanHex, trace.FlagsSampled)
	out := logLine(ctx, t)

	if got := out[gcpTraceKey]; got != testTraceHex {
		t.Errorf("%s = %v, want bare trace id %q (no project prefix)", gcpTraceKey, got, testTraceHex)
	}
	if got := out[gcpSpanIDKey]; got != testSpanHex {
		t.Errorf("%s = %v, want %q", gcpSpanIDKey, got, testSpanHex)
	}
	if sampled, ok := out[gcpTraceSampledKey].(bool); !ok || !sampled {
		t.Errorf("%s = %v (bool=%v), want true", gcpTraceSampledKey, out[gcpTraceSampledKey], ok)
	}
}

func TestTraceHookNotSampled(t *testing.T) {
	ctx := newSpanContext(t, testTraceHex, testSpanHex, 0)
	out := logLine(ctx, t)

	if _, ok := out[gcpTraceKey]; !ok {
		t.Errorf("%s should be present even when not sampled", gcpTraceKey)
	}
	// trace_sampled is omitted when the span is not sampled, matching the golden path.
	if _, ok := out[gcpTraceSampledKey]; ok {
		t.Errorf("%s must be omitted when not sampled", gcpTraceSampledKey)
	}
}

func TestTraceHookNoSpan(t *testing.T) {
	out := logLine(context.Background(), t)

	for _, k := range []string{gcpTraceKey, gcpSpanIDKey, gcpTraceSampledKey} {
		if _, ok := out[k]; ok {
			t.Errorf("%s must not be set when the context has no valid span", k)
		}
	}
}

func TestWithNoTraceDisablesHook(t *testing.T) {
	ctx := newSpanContext(t, testTraceHex, testSpanHex, trace.FlagsSampled)
	out := logLine(ctx, t, WithNoTrace())

	if _, ok := out[gcpTraceKey]; ok {
		t.Errorf("%s must not be set when WithNoTrace() is used", gcpTraceKey)
	}
}

// The slog handler is a separate code path that must thread the context through
// to the trace hook. slog only forwards the context via the *Context variants.
func TestSlogHandlerAddsTraceFields(t *testing.T) {
	ctx := newSpanContext(t, testTraceHex, testSpanHex, trace.FlagsSampled)

	var buf bytes.Buffer
	h := NewSlogHandler(WithWriter(&buf), WithLevel(InfoLevel))
	logger := slog.New(h)
	logger.InfoContext(ctx, "hello")

	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal %q: %v", buf.String(), err)
	}

	if got := out[gcpTraceKey]; got != testTraceHex {
		t.Errorf("%s = %v, want %q", gcpTraceKey, got, testTraceHex)
	}
	if got := out[gcpSpanIDKey]; got != testSpanHex {
		t.Errorf("%s = %v, want %q", gcpSpanIDKey, got, testSpanHex)
	}
	if sampled, ok := out[gcpTraceSampledKey].(bool); !ok || !sampled {
		t.Errorf("%s = %v (bool=%v), want true", gcpTraceSampledKey, out[gcpTraceSampledKey], ok)
	}
}

func TestSlogHandlerWithNoTrace(t *testing.T) {
	ctx := newSpanContext(t, testTraceHex, testSpanHex, trace.FlagsSampled)

	var buf bytes.Buffer
	h := NewSlogHandler(WithWriter(&buf), WithLevel(InfoLevel), WithNoTrace())
	logger := slog.New(h)
	logger.InfoContext(ctx, "hello")

	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal %q: %v", buf.String(), err)
	}

	if _, ok := out[gcpTraceKey]; ok {
		t.Errorf("%s must not be set when WithNoTrace() is used", gcpTraceKey)
	}
}
