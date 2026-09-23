package logging

import (
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

// GCP Cloud Logging promotes these special JSON fields out of the jsonPayload
// and into the LogEntry itself, which is what links a log line to its trace in
// the Logs Explorer and the Cloud Trace integration.
//
// The trace field is written as the bare trace id. That is the preferred format;
// the projects/[PROJECT-ID]/traces/[TRACE-ID] resource name is a legacy format,
// so we intentionally do not require or reference a project id here.
//
// See:
//   - https://cloud.google.com/logging/docs/agent/logging/configuration#special-fields
//   - https://cloud.google.com/trace/docs/trace-log-integration
const (
	gcpTraceKey        = "logging.googleapis.com/trace"
	gcpSpanIDKey       = "logging.googleapis.com/spanId"
	gcpTraceSampledKey = "logging.googleapis.com/trace_sampled"
)

// gcpTraceHook enriches each event with the Google Cloud trace/span fields taken
// from the OpenTelemetry span context carried on the event's context.
//
// It is a no-op unless the event carries a context with a valid span, so it is
// safe to attach to every logger. Attach a context to an event with Ctx(ctx),
// e.g. logging.Ctx(ctx).Info() or logging.Info().Ctx(ctx). The slog handler wires
// the context through automatically.
type gcpTraceHook struct{}

func (gcpTraceHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	sc := trace.SpanContextFromContext(e.GetCtx())
	if !sc.IsValid() {
		return
	}

	e.Str(gcpTraceKey, sc.TraceID().String())
	e.Str(gcpSpanIDKey, sc.SpanID().String())

	// Match the golden path: only emit trace_sampled when the span is sampled.
	if sc.IsSampled() {
		e.Bool(gcpTraceSampledKey, true)
	}
}
