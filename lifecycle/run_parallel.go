package lifecycle

import (
	"context"
	"errors"
)

type Runnable interface {
	Run(context.Context) error
}

type Runnables map[string]Runnable

type RunFn = func(context.Context) error
type RunFns map[string]RunFn

type RunTermination struct {
	Name string
	Err  error
}

// NOTE: Returns true if other runs should be stopped
type TerminationHandlerFn func(*RunTermination) bool

func RunParallel(
	ctx context.Context,
	runners map[string]Runnable,
	terminationHandler TerminationHandlerFn,
) error {
	fns := make(RunFns, len(runners))

	for k, runable := range runners {
		fns[k] = runable.Run
	}

	return RunParallelFn(ctx, fns, terminationHandler)
}

func RunParallelFn(
	ctx context.Context,
	runners RunFns,
	terminationHandler TerminationHandlerFn,
) error {
	if len(runners) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	termCh := make(chan RunTermination)

	for name, run := range runners {
		go func() {
			termCh <- RunTermination{
				Name: name,
				Err:  run(ctx),
			}
		}()
	}

	var err error
	nterminated := 0

	for nterminated < len(runners) {
		term := <-termCh
		nterminated += 1

		shouldAbort := terminationHandler(&term)
		if shouldAbort {
			cancel()
		}

		err = errors.Join(err, term.Err)
	}

	return err
}
