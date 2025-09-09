package logging

import (
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// -----------------------
// Logging <-> Zerolog aliases
// -----------------------

type (
	// Types
	ConsoleWriter = zerolog.ConsoleWriter
	Level         = zerolog.Level
	Logger        = zerolog.Logger
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

// -----------------------
// Initialize and Prepare logging library for GCP
// -----------------------

func init() {
	// Configure zerolog for GCP logging
	zerolog.LevelFieldName = "severity"
	zerolog.LevelWarnValue = "warning"
	zerolog.TimestampFieldName = "timestamp"
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.ErrorStackMarshaler = marshalStack

	// Get log level
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "warning" // Default log level is set to warning as per the ADR spec
	}

	// NOTE:
	// We set the global zerolog level by overriding the default global logger in zerolog.
	// Do not call zerolog.SetGlobalLevel(), as that will make it impossible to raise the log level locally in other loggers!!
	logger := log.Logger.Level(convertStrToZLogLevel(level))
	logger = logger.With().Stack().Logger()
	log.Logger = logger
}

// -----------------------
// Logging Configuration
// -----------------------

type Config struct {
	w            io.Writer
	level        *zerolog.Level
	caller       bool
	noStackTrace bool
	noTimestamp  bool
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

func WithNoStackTrace() Option {
	return func(c *Config) {
		c.noStackTrace = true
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

func WithCaller() Option {
	return func(c *Config) {
		c.caller = true
	}
}

func WithLevel(level Level) Option {
	return func(c *Config) {
		c.level = &level
	}
}

func New(opts ...Option) Logger {
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
	if !cfg.noStackTrace {
		ctx = ctx.Stack()
	}
	if cfg.caller {
		ctx = ctx.Caller()
	}

	logger := ctx.Logger()
	if cfg.level != nil {
		logger = logger.Level(*cfg.level)
	}

	return logger
}

func NewConsoleWriter(opts ...Option) ConsoleWriter {
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
		level:       convertZLogLevelToSLog(logger.GetLevel()),
		noTimestamp: cfg.noTimestamp,
	}
}
