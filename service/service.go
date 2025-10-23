package service

import (
	"context"
	"errors"
	"os"

	"github.com/yandzee/go-svc/lifecycle"
)

type ControllableInstance interface {
	lifecycle.Runnable

	Prepare(context.Context) error
	Shutdown(context.Context) error
}

func Start(ctx context.Context, instance ControllableInstance) {
	host := &Host{
		Instance: instance,
	}

	if err := host.Prepare(ctx); err != nil {
		ExitOnError(err, 1)
	}

	if err := host.Run(); err != nil {
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
