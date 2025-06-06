package logging_test

import (
	"log/slog"

	"github.com/entur/go-logging"
)

func Example() {
	// Default global logger
	// logging.Info().Msg("Hello from the global logger!")

	// New child logger
	// logger := logging.New()
	// err := fmt.Errorf("oh no, an error")
	// logger.Error().Err(err).Msg("An internal error occurred")

	// New child logger with stacktraced error
	// logger := logging.New()
	// err := logging.NewStackTraceError("oh no, an error")
	// logger.Error().Stack().Err(err).Msg("An internal error occurred!")

	// Child logger with custom writer
	w := logging.NewConsoleWriter(logging.WithNoColor(), logging.WithNoTimestamp())

	logger := logging.New(logging.WithWriter(w), logging.WithLevel(logging.InfoLevel))
	logger.Info().Msg("Hello from my new child logger!")

	// Child logger with custom writer and level which won't be logged
	logger = logging.New(logging.WithWriter(w), logging.WithLevel(logging.WarnLevel))
	logger.Info().Msg("Hello from my new custom child logger!")

	// Slog logger with zerolog handler. Inefficient, so should only be used sparingly
	// if some other SDK is able to take a custom slog handler.
	h := logging.NewSlogHandler(logging.WithWriter(w))
	slogger := slog.New(h)
	slogger = slogger.With(
		slog.Int("some_int", 1),
		slog.String("some_string", "huh"),
		slog.Group("some_group",
			slog.Int("some_nested_int", 3),
			slog.Group("some_nested_group",
				slog.Float64("some_nested_nested_float", 100.0),
			),
		),
	)
	slogger.Info("Hello from my new custom slog handler")

	// Output:
	// INF Hello from my new child logger!
	// INF Hello from my new custom slog handler some_group={"some_nested_group":{"some_nested_nested_float":100},"some_nested_int":3} some_int=1 some_string=huh
}
