# go-logging

A simple SDK for standardized GCP logging in golang. 

## Start using the SDK

You need to enable private go modules from entur:

```sh
go env -w GOPRIVATE='github.com/entur/*'
env GIT_TERMINAL_PROMPT=1 go get github.com/entur/go-logging # to fix if you default to https
# git config --global --add url."git@github.com:".insteadOf "https://github.com/" # if you want ssh default always
```

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
