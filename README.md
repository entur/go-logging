# go-logging

## Start using the SDK

The default log level will be derived from the `LOG_LEVEL` or `COMMON_ENV` environment variables if defined, or set to `trace` if not. The former values take priority over the latter. Valid values are:
* `fatal`
* `panic`
* `error`
* `warning`
* `info`
* `debug`
* `trace`

## Minimal example

See `./logging_test.go` for a complete test.
