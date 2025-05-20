package logging_test

import (
	"github.com/entur/go-logging"
	"github.com/rs/zerolog"
)

func Example() {
	// Default global logger
	// logging.Info().Msg("Hello from the global logger!")

	// Default child logger
	// logger := logging.New()
	// err := fmt.Errorf("oh no, an error")
	// logger.Error().Err(err).Msg("An internal error occurred")

	// Default child logger with stacktraced error
	// logger := logging.New()
	// err := logging.NewStackTraceError("oh no, an error")
	// logger.Error().Stack().Err(err).Msg("An internal error occurred!")

	// Logger with custom writer
	w := zerolog.NewConsoleWriter()
	w.NoColor = true
	w.PartsExclude = []string{"timestamp"}

	logger := logging.New(logging.WithWriter(w))
	logger.Info().Msg("Hello from my new child logger!")

	// Logger with custom writer and level
	logger = logging.New(logging.WithWriter(w), logging.WithLevel(zerolog.WarnLevel))
	logger.Info().Msg("Hello from my new custom child logger!")

	// Output:
	// INF Hello from my new child logger!
}
