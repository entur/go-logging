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

## Tests
This project makes use of Example tests. To run them, simply use use the following command
```sh
go test ./...
```

## Examples
Interested in seeing how Go-Logging is used in practice at Entur? Take a look at the following repositories:
* [https://github.com/entur/go-orchestrator](https://github.com/entur/go-orchestrator)