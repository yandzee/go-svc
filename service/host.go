package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/yandzee/go-svc/log"
)

const DefaultTerminationTimeout = 5 * time.Second

type Host struct {
	Instance           ControllableInstance
	Log                *slog.Logger
	TerminationTimeout time.Duration

	shutdownCh   chan struct{}
	shutdownOnce sync.Once
}

func (h *Host) Prepare(ctx context.Context) error {
	if h.Instance == nil {
		return fmt.Errorf("service.Host: Instance field must be set")
	}

	h.Log = log.OrDiscard(h.Log)

	if h.TerminationTimeout <= 0 {
		h.TerminationTimeout = DefaultTerminationTimeout
	}

	if err := h.Instance.Prepare(ctx); err != nil {
		return err
	}

	return nil
}

func (h *Host) Run(ctx context.Context) error {
	if h.Instance == nil {
		return fmt.Errorf("instance is not set")
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	h.shutdownCh = make(chan struct{})
	h.shutdownOnce = sync.Once{}

	ctx, cancel := context.WithCancel(ctx)
	signalCh := h.setupSignalHandling(ctx, &wg)

	var instanceErr error
	instanceReturned := false
	errCh := make(chan error)

	// NOTE: We are not sure if instance.Run is going to return timely...
	go func() {
		errCh <- h.Instance.Run(ctx)
	}()

	select {
	case err := <-errCh:
		instanceReturned = true
		instanceErr = err
	case <-signalCh:
	case <-h.shutdownCh:
	case <-ctx.Done():
	}

	// NOTE: ...so we notify everyone that music is about to stop.
	cancel()

	// NOTE: Wait for the signal goroutine to return
	wg.Wait()

	if instanceReturned {
		return instanceErr
	}

	h.Log.Warn(
		"Waiting instance to return before forced return",
		"delay", h.TerminationTimeout.String(),
	)

	select {
	case err := <-errCh:
		return err
	case <-time.After(h.TerminationTimeout):
	}

	return ctx.Err()
}

func (h *Host) Shutdown(ctx context.Context) error {
	if h.Instance == nil || h.shutdownCh == nil {
		return nil
	}

	h.shutdownOnce.Do(func() {
		close(h.shutdownCh)
	})

	return nil
}

func (h *Host) setupSignalHandling(ctx context.Context, wg *sync.WaitGroup) chan struct{} {
	ch := make(chan os.Signal, 32)
	signalCh := make(chan struct{})

	go func() {
		select {
		case sig := <-ch:
			h.Log.Warn("os signal received", "sig", sig.String())
		case <-ctx.Done():
			h.Log.Warn("os signal monitoring done", "err", ctx.Err())
		}

		signal.Stop(ch)
		close(signalCh)

		h.Log.Debug("exitting signal handler loop")
		wg.Done()
	}()

	signal.Notify(
		ch,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	return signalCh
}
