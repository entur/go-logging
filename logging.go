package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// Functions
	Fatal = log.Fatal
	Panic = log.Panic
	Error = log.Error
	Warn  = log.Warn
	Info  = log.Info
	Debug = log.Debug
	Trace = log.Trace
	Ctx   = zerolog.Ctx
	// Levels
	FatalLevel = zerolog.FatalLevel
	PanicLevel = zerolog.PanicLevel
	ErrorLevel = zerolog.ErrorLevel
	WarnLevel  = zerolog.WarnLevel
	InfoLevel  = zerolog.InfoLevel
	DebugLevel = zerolog.DebugLevel
	TraceLevel = zerolog.TraceLevel
	NoLevel    = zerolog.NoLevel
	Disabled   = zerolog.Disabled
)

func setLevel(level string) {
	switch strings.ToLower(level) {
	case "fatal":
		zerolog.SetGlobalLevel(FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(PanicLevel)
	case "error":
		zerolog.SetGlobalLevel(ErrorLevel)
	case "warning":
		zerolog.SetGlobalLevel(WarnLevel)
	case "info":
		zerolog.SetGlobalLevel(InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(DebugLevel)
	default:
		zerolog.SetGlobalLevel(TraceLevel)
	}
}

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

	setLevel(level)
}

type Config struct {
	w           io.Writer
	level       *zerolog.Level
	noTimestamp bool
	// ConsoleWriter
	noColor       bool
	fieldsExclude []string
}

type Option func(*Config)

func WithWriter(writers ...io.Writer) Option {
	var w io.Writer

	num := len(writers)
	if num > 0 {
		if num > 1 {
			w = zerolog.MultiLevelWriter(writers...)
		} else {
			w = writers[0]
		}
	}

	return func(c *Config) {
		c.w = w
	}
}

func WithNoTimestamp() Option {
	return func(c *Config) {
		c.noTimestamp = true
	}
}

func WithNoColor() Option {
	return func(c *Config) {
		c.noColor = true
	}
}

func WithExcludeFields(fields ...string) Option {
	return func(c *Config) {
		c.fieldsExclude = fields
	}
}

func WithLevel(level zerolog.Level) Option {
	return func(c *Config) {
		c.level = &level
	}
}

func New(opts ...Option) zerolog.Logger {
	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}

	w := cfg.w
	if w == nil {
		w = os.Stderr
	}

	ctx := zerolog.New(w).With()
	if !cfg.noTimestamp {
		ctx = ctx.Timestamp()
	}

	logger := ctx.Logger()
	if cfg.level != nil {
		logger = logger.Level(*cfg.level)
	}

	return logger
}

func NewConsoleWriter(opts ...Option) zerolog.ConsoleWriter {
	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}

	f := func(w *zerolog.ConsoleWriter) {
		w.NoColor = cfg.noColor

		if cfg.noTimestamp {
			w.PartsExclude = []string{"timestamp"}
		}

		if len(cfg.fieldsExclude) > 0 {
			w.FieldsExclude = cfg.fieldsExclude
		}
	}

	return zerolog.NewConsoleWriter(f)
}

func NewSlogHandler(opts ...Option) slog.Handler {
	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}

	w := cfg.w
	if w == nil {
		w = os.Stderr
	}

	logger := zerolog.New(w)
	if cfg.level != nil {
		logger = logger.Level(*cfg.level)
	}

	return &SLogHandler{
		logger:      &logger,
		level:       levelZlogToSlog(logger.GetLevel()),
		noTimestamp: cfg.noTimestamp,
	}
}
