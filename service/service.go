package service

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/yandzee/go-svc/lifecycle"
	"github.com/yandzee/go-svc/log"
)

type ControllableInstance interface {
	lifecycle.Runnable

	Prepare(context.Context) error
	Shutdown(context.Context) error
}

type HostableService[C any] interface {
	ControllableInstance

	SetLogger(*slog.Logger)
	SetConfig(C)
}

type ServiceConfig interface {
	LogOptions() log.LoggerOptions
	LogRecords() []slog.Record
}

func Start[C any](ctx context.Context, svc HostableService[C], cfg C) {
	var logger *slog.Logger

	if cfg, ok := any(cfg).(ServiceConfig); ok {
		logger = log.Init(cfg.LogOptions())

		hasFatal := false
		for _, logRecord := range cfg.LogRecords() {
			if err := logger.Handler().Handle(ctx, logRecord); err != nil {
				panic(err.Error())
			}

			hasFatal = hasFatal || logRecord.Level == slog.LevelError
		}

		if hasFatal {
			ExitOnError(errors.New("start: config has errors"), 1)
		}
	} else {
		logger = log.Init(log.LoggerOptions{
			Level:        slog.LevelDebug,
			IsStructured: false,
			IsColored:    true,
		})
	}

	svc.SetLogger(logger)
	svc.SetConfig(cfg)

	StartInstance(ctx, &Host{
		Instance: svc,
		Log:      logger.With("log", "service.Host"),
	})
}

func StartInstance(ctx context.Context, instance ControllableInstance) {
	if err := instance.Prepare(ctx); err != nil {
		ExitOnError(err, 1)
	}

	if err := instance.Run(ctx); err != nil {
		ExitOnError(err, 2)
	}

	ExitOnError(nil, 0)
}

func ExitOnError(err error, errCode ...int) {
	exitCode := 0
	if len(errCode) > 0 {
		exitCode = errCode[0]
	}

	switch {
	case err == nil:
		fallthrough
	case errors.Is(err, context.Canceled):
		os.Exit(0)
	default:
		os.Exit(exitCode)
	}
}
