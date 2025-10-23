package lifecycle

import (
	"context"
	"errors"
)

type Runnable interface {
	Run(context.Context) error
}

type Runnables map[string]Runnable

type RunTermination struct {
	Name string
	Err  error
}

func RunParallel(
	ctx context.Context,
	runners map[string]Runnable,
	terminationHandler func(*RunTermination) bool,
) error {
	if len(runners) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	termCh := make(chan RunTermination)

	for name, runnable := range runners {
		go func() {
			termCh <- RunTermination{
				Name: name,
				Err:  runnable.Run(ctx),
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
			return term.Err
		}

		if term.Err == nil {
			continue
		}

		err = errors.Join(err, term.Err)
	}

	return nil
}
