package logging_test

import (
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

	logger := logging.New(logging.WithWriter(w))
	logger.Info().Msg("Hello from my new child logger!")

	// Child logger with custom writer and level which won't be logged
	logger = logging.New(logging.WithWriter(w), logging.WithLevel(logging.WarnLevel))
	logger.Info().Msg("Hello from my new custom child logger!")

	// Output:
	// INF Hello from my new child logger!
}
