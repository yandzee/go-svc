package log

import (
	"log/slog"
	"os"
	"time"

	"github.com/rs/zerolog"
	slogzerolog "github.com/samber/slog-zerolog/v2"
)

func Init(lvl slog.Level, isProduction bool) *slog.Logger {
	zerologLogger := zerolog.New(os.Stdout)

	if !isProduction {
		zerologLogger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.DateTime,
		})
	}

	slogHandler := slogzerolog.Option{
		Level:     lvl,
		Logger:    &zerologLogger,
		AddSource: false,
	}.NewZerologHandler()

	logger := slog.New(slogHandler)

	return logger
}
