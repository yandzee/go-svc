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

type Host struct {
	Instance ControllableInstance
	Log      *slog.Logger

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func (h *Host) Prepare(_ctx context.Context) error {
	if h.Instance == nil {
		return fmt.Errorf("service.Host: Instance field must be set")
	}

	ctx, cancel := context.WithCancel(_ctx)

	h.ctx = ctx
	h.cancel = cancel

	if err := h.setupSignalHandlers(); err != nil {
		return err
	}

	if err := h.Instance.Prepare(ctx); err != nil {
		return err
	}

	return nil
}

func (h *Host) Run() error {
	if h.Instance == nil {
		return fmt.Errorf("nothing to run")
	}

	if h.ctx == nil {
		return fmt.Errorf("host.Prepare() must be called first")
	}

	err := h.Instance.Run(h.ctx)

	// This .Wait() prevents goroutine (maybe main) from exiting before other
	// control goroutines do
	h.wg.Wait()

	return err
}

func (h *Host) setupSignalHandlers() error {
	c := make(chan os.Signal, 32)
	signal.Notify(
		c,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go func() {
		defer signal.Stop(c)
		defer h.log().Debug("exitting signal handler loop")

	F:
		for {
			select {
			case sig := <-c:

				if shouldBreak := h.handleSignal(sig); shouldBreak {
					break F
				}
			case <-h.ctx.Done():
				break F
			}
		}
	}()

	return h.ctx.Err()
}

func (h *Host) handleSignal(sig os.Signal) bool {
	h.wg.Add(1)

	log := h.log().With("sig", sig.String())
	log.Warn("handling received OS signal")

	log.Warn("cancelling root context")
	h.cancel()

	log.Warn("running graceful shutdown on ServiceHost impl")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Inner.Shutdown() will switch Goroutine to the main thread, which after
	// exitting terminates all remaining goroutines, including one where
	// Inner.Shutdown is running, so we have to block the execution artificially
	defer h.wg.Done()

	err := h.Instance.Shutdown(ctx)
	attrs := []any{}
	lvl := slog.LevelInfo

	if err != nil {
		attrs = append(attrs, slog.String("err", err.Error()))
		lvl = slog.LevelError
	}

	log.Log(ctx, lvl, "graceful shutdown finished", attrs...)
	return true
}

func (h *Host) log() *slog.Logger {
	if h.Log != nil {
		return h.Log
	}

	h.Log = log.Discard()
	return h.Log
}
