# go-logging

Go-Logging is intended as a simple-to-use SDK for high-performance logging in GCP and locally. It supports caller location identification, optional stacktraces, colorful logging and more.

## Quickstart

### Install 
```sh
go get github.com/entur/go-logging
go mod tidy
```

### Basic Usage 
```golang
import (
  "github.com/entur/go-logging"
)

func main() {
  // Global logger
  logging.Warn().Msg("Starting up my application")

  // Formatted logging
  logging.Warn().Msgf("Starting my %s", "application")

  // Local logger
  logger := logging.New()
  logger.Warn().Msg("Starting up my application")

  // Local logger with custom level
  logger = logging.New(
    logging.WithLevel(logging.DebugLevel)
  )
  logger.Debug().Msg("Starting up my application")
}
```

## Overview

### Log Levels
The default log level will be derived from the `LOG_LEVEL` environment variable if defined, or set to the `warning` level if not as per Entur's [Architecture Descision Record](https://enturas.atlassian.net/wiki/spaces/eat/pages/5318344894/2022-10-31+All+services+must+have+a+balanced+log+level). 

Valid environment variable values for log levels are:
* `fatal` `ftl`
* `panic` `pnc`
* `error` `err`
* `warning` `wrn`
* `info` `inf`
* `debug` `dbg`
* `trace` `trc`

Log levels can also be specified for on logger instance creation like so:
```go
func main() {
  // New logger with its own level
  logger = logging.New(
    logging.WithLevel(logging.DebugLevel)
  )

  // Child logger with its own level again
  additionalLogger := logger.Level(logging.TraceLevel)
}
```

### Stack Traces
The Go-Logging SDK supports logging of errors with stacktraces. To do so, simply create a new or wrap an existing error using the `logging.NewStackTraceError()` before you dispatch it for logging.

```go
import (
  "fmt"

  "github.com/entur/go-logging"
)

func newErr() error {
  err := logging.NewStackTraceError("called newErr()")
  return err
}

func wrappedErr() error {
  existingErr := fmt.Errorf("called wrappedErr()")
  err := logging.NewStackTraceError("%w", existingErr) // Stack traces will be retrieved at the point NewStackTraceError is called
  return err
}

func main() {
  err := newErr()
  logging.Error().Err(err).Msg("An internal error occurred")

  err = wrappedErr()
  logging.Error().Err(err).Msg("An internal error occurred")
}
```

### Calling Location Info
The Go-Logging SDK will automatically include the caller info when you use the global logging functions, or create a new instanced logger. If you want to disable the feature, you can prove the `WithNoCaller()` option to the logging constructor.

```go
import (
  "github.com/entur/go-logging"
)

func main() {
  // Disable caller info in child logger
  logger := logging.New(
    logging.WithNoCaller()
  )

  logger.Debug().Msg("This log won't include caller info")

  // You can still include caller info individually, even if it is disabled by default
  logger.Debug().Caller("This log will include caller info")
}
```

### Trace Correlation
When a log is emitted with a `context.Context` that carries a valid [OpenTelemetry](https://opentelemetry.io/) span, the Go-Logging SDK automatically enriches the log with the special fields Google Cloud Logging uses to link a log line to its trace:

* `logging.googleapis.com/trace` — the bare trace id (the preferred format; the legacy `projects/[PROJECT-ID]/traces/[TRACE-ID]` resource name is not used, so no project id is required)
* `logging.googleapis.com/spanId` — the span id
* `logging.googleapis.com/trace_sampled` — set to `true` only when the span is sampled

This makes logs and traces correlate automatically in the Logs Explorer and the Cloud Trace integration, matching the golden path used by the JVM services via [entur/cloud-logging](https://github.com/entur/cloud-logging).

Attach the context to the log event so the SDK can read the span from it:

```go
import (
  "context"

  "github.com/entur/go-logging"
)

func handle(ctx context.Context) {
  // ctx carries the current OpenTelemetry span (e.g. from otelhttp middleware)
  logging.Ctx(ctx).Info().Msg("handling request")

  // equivalently
  logging.Info().Ctx(ctx).Msg("handling request")
}
```

The enrichment is a no-op when the context has no valid span, so it is always safe to leave on. To disable it, pass the `WithNoTrace()` option to the logging constructor:

```go
logger := logging.New(
  logging.WithNoTrace(),
)
```

The `slog` handler returned by `NewSlogHandler()` wires the context through automatically, so trace correlation works there too.

#### Configuring tracing in your service
Go-Logging only *reads* the span from the context — it never decides whether a request is traced or sampled. Those decisions belong to the OpenTelemetry `TracerProvider` you set up once at service start-up. A minimal setup with the Cloud Trace exporter looks like this:

```go
import (
  texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
  "go.opentelemetry.io/otel"
  "go.opentelemetry.io/otel/propagation"
  sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func initTracing(ctx context.Context, projectID string) (func(context.Context) error, error) {
  exp, err := texporter.New(texporter.WithProjectID(projectID))
  if err != nil {
    return nil, err
  }

  tp := sdktrace.NewTracerProvider(
    sdktrace.WithBatcher(exp),
    sdktrace.WithSampler(sdktrace.AlwaysSample()), // the sampling decision
  )
  otel.SetTracerProvider(tp)

  // Propagate trace context in/out of the service. The GCP one-way propagator
  // continues a trace started by the Cloud Run load balancer; W3C is the modern
  // default for outbound calls.
  otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
    propagation.TraceContext{},
    propagation.Baggage{},
  ))

  return tp.Shutdown, nil
}
```

Wrap your HTTP handlers with [`otelhttp`](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp) (or the equivalent gRPC interceptor) so each request gets a span placed on its `context.Context`. Go-Logging then picks that span up automatically.

##### Sampling
The sampler passed to `WithSampler(...)` controls the `logging.googleapis.com/trace_sampled` value Go-Logging emits:

| Sampler | Effect |
| --- | --- |
| `sdktrace.AlwaysSample()` | every request is sampled — `trace_sampled: true` on every log with a span |
| `sdktrace.NeverSample()` | nothing is sampled — the `trace_sampled` field is never emitted |
| `sdktrace.TraceIDRatioBased(0.01)` | probabilistic — roughly 1% of traces sampled |
| `sdktrace.ParentBased(sampler)` | inherit the sampled bit from the incoming request's trace context |

> **Note:** `ParentBased` inherits the caller's decision. Behind the Cloud Run load balancer that marks only ~0.1% of requests as sampled, so a service using `ParentBased(AlwaysSample())` will trace almost nothing. Use `AlwaysSample()` (or a ratio sampler) when the service should make its own decision. Trace continuity across services is handled by the propagator and is independent of the sampler, so cross-service traces stay linked regardless.

## Tests
This project makes use of Example tests. To run them, simply use use the following command
```sh
go test ./...
```

## Examples
Interested in seeing how Go-Logging is used in practice at Entur? Take a look at the following repositories:
* [https://github.com/entur/go-orchestrator](https://github.com/entur/go-orchestrator)