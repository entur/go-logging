package logging

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	// Configure zerolog for GCP logging
	zerolog.LevelFieldName = "severity"
	zerolog.LevelWarnValue = "warning"
	zerolog.TimestampFieldName = "timestamp"
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.ErrorStackMarshaler = marshalStack

	// Set log level
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		// If no default LOG_LEVEL value is set, try to derive it from our runtime environment.
		// * See ADR https://enturas.atlassian.net/wiki/spaces/eat/pages/5318344894/2022-10-31+All+services+must+have+a+balanced+log+level
		// * See COMMON_ENV in Helm Chart https://github.com/entur/helm-charts
		env := os.Getenv("COMMON_ENV")
		switch strings.ToLower(env) {
		case "dev":
			// NO DEFAULT SPECIFIED IN ADR
		case "tst":
			// NO DEFAULT SPECIFIED IN ADR
		case "prd":
			level = "warning"
		}
	}

	switch strings.ToLower(level) {
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warning":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}
}

type Config struct {
	w          io.Writer
	level      *zerolog.Level
}

type Option func(*Config)

func WithWriter(w io.Writer) Option {
	return func(c *Config) {
		c.w = w
	}
}

func WithLevel(level zerolog.Level) Option {
	return func(c *Config) {
		c.level = &level
	}
}

func New(options ...Option) zerolog.Logger {
	cfg := &Config{}
	for _, opt := range options {
		opt(cfg)
	}

	logger := log.Logger
	if cfg.w != nil {
		logger = log.Output(cfg.w)
	}
	if cfg.level != nil {
		logger = logger.Level(*cfg.level)
	}

	return logger.With().Logger()
}
