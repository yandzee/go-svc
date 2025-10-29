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

type TerminationContext struct {
	Name          string
	Err           error
	CancelContext context.CancelFunc
}

// NOTE: Returns true if other runs should be stopped
type TerminationHandlerFn func(*TerminationContext)

func RunParallel(
	ctx context.Context,
	runners map[string]Runnable,
	terminationHandler ...TerminationHandlerFn,
) error {
	fns := make(RunFns, len(runners))

	for k, runnable := range runners {
		fns[k] = runnable.Run
	}

	return RunParallelFn(ctx, fns, terminationHandler...)
}

func RunParallelFn(
	ctx context.Context,
	runners RunFns,
	terminationHandler ...TerminationHandlerFn,
) error {
	if len(runners) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	termCh := make(chan TerminationContext)

	for name, run := range runners {
		go func() {
			termCh <- TerminationContext{
				Name:          name,
				Err:           run(ctx),
				CancelContext: cancel,
			}
		}()
	}

	var err error
	nterminated := 0

	for nterminated < len(runners) {
		term := <-termCh
		nterminated += 1

		if len(terminationHandler) > 0 {
			terminationHandler[0](&term)
		}

		err = errors.Join(err, term.Err)
	}

	return err
}
