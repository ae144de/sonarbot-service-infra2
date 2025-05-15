package log

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// NewLogger configures zerolog to write to console with human-friendly output when appropriate.
func NewLogger() zerolog.Logger {
	// Human-friendly only if stdout is a TTY
	out := zerolog.ConsoleWriter{Out: os.Stdout}
	logger := log.Output(out).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	return logger
}
