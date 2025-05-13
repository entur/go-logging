package logging_test

import (
	"github.com/entur/go-logging"
	"github.com/rs/zerolog"
)

func Example() {
	// Default logger
	//logger := New()
	//err := fmt.Errorf("oh no, an error")
	//logger.Error().Err(err).Msg("An internal error occured")

	// Logger with custom writer
	w := zerolog.NewConsoleWriter()
	w.NoColor = true
	w.PartsExclude = []string{"timestamp"}

	logger := logging.New(logging.WithWriter(w))
	logger.Info().Msg("Hello!")

	// Logger with custom writer and level
	logger = logging.New(logging.WithWriter(w), logging.WithLevel(zerolog.WarnLevel))
	logger.Info().Msg("Hello!")

	// Output:
	// INF Hello!
}
