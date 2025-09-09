# go-logging

Go-Loging is intended as a simple-to-use SDK for high-performance logging in GCP and locally. It supports caller location identification, optional stacktraces, colorful logging and more.

## Quickstart

### Install 
```sh
go get github.com/entur/go-logging
go mod tidy
```

### Import 
```golang
import (
  "github.com/entur/go-logging"
)
```

### Basic Usage 
```golang
import (
  "github.com/entur/go-logging"
)

func main() {
  // Global logger
  logging.Warn().Msg("Starting up my application")

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
When logging similar errors and messages at multiple locations, it can be useful to have the ability to identify the original file and call location. In the Go-Logging SDK this can be done by just calling the `Caller()` function on a logging event before dispatching it for sending, or adding it as a default setting to an instanced logger.

```go
import (
  "github.com/entur/go-logging"
)

func main() {
  // When global logging
  logging.Debug().Caller().Msg("My location is here!")

  // Setup local child logger to include caller by-default
    logger = logging.New(
      logging.WithCaller()
    )
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