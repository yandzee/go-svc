package log

import (
	"log/slog"
	"os"
	"time"

	"github.com/rs/zerolog"
	slogzerolog "github.com/samber/slog-zerolog/v2"
	"github.com/yandzee/go-svc/crypto"
)

type LoggerOptions struct {
	Level        slog.Level
	IsStructured bool
	IsColored    bool
}

func Init(opts LoggerOptions) *slog.Logger {
	zerologLogger := zerolog.New(os.Stdout)

	if !opts.IsStructured {
		zerologLogger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.DateTime,
			NoColor:    !opts.IsColored,
		}).Level(intoZerologLevel(opts.Level))
	}

	slogHandler := slogzerolog.Option{
		Level:     opts.Level,
		Logger:    &zerologLogger,
		AddSource: false,
	}.NewZerologHandler()

	instanceId := crypto.RandomHex(4)
	logger := slog.New(slogHandler).With("iid", instanceId)

	return logger
}

func intoZerologLevel(lvl slog.Level) zerolog.Level {
	switch lvl {
	case slog.LevelDebug:
		return zerolog.DebugLevel
	case slog.LevelInfo:
		return zerolog.InfoLevel
	case slog.LevelWarn:
		return zerolog.WarnLevel
	case slog.LevelError:
		return zerolog.ErrorLevel
	}

	return zerolog.DebugLevel
}
