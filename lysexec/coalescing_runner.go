package lysexec

import (
	"context"
	"log"
	"log/slog"
	"sync"
	"time"
)

// CoalescingRunner manages the execution of a function with many potential triggers.
// It ensures that the function is not run concurrently and handles debouncing of triggers.
type CoalescingRunner struct {
	ctx              context.Context
	runFunc          func(context.Context) error
	runFuncStr       string        // for logging
	debounceDuration time.Duration // duration to wait before running the function after a trigger, e.g. 100ms
	logger           *slog.Logger

	mu      sync.Mutex // guards pending and running
	pending bool
	running bool
	wg      sync.WaitGroup // ensures that the function can be waited on for graceful shutdown
}

// NewCoalescingRunner creates a new CoalescingRunner.
func NewCoalescingRunner(
	ctx context.Context,
	runFunc func(ctx context.Context) error,
	runFuncStr string,
	debounceDuration time.Duration,
	logger *slog.Logger,
) *CoalescingRunner {

	if runFunc == nil {
		log.Fatal("runFunc is required")
	}
	if runFuncStr == "" {
		log.Fatal("runFuncStr is required")
	}
	if debounceDuration <= 0 {
		log.Fatal("debounceDuration must be > 0")
	}
	if logger == nil {
		log.Fatal("logger is required")
	}

	return &CoalescingRunner{
		ctx:              ctx,
		runFunc:          runFunc,
		runFuncStr:       runFuncStr,
		debounceDuration: debounceDuration,
		logger:           logger,
	}
}

// Trigger sets the pending flag and starts the function in a goroutine if it's not already running.
func (r *CoalescingRunner) Trigger() {

	// ensure no action is taken if the context is already canceled
	if r.ctx.Err() != nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// set pending true to indicate that another run is needed
	r.pending = true

	// exit if already running
	if r.running {
		return
	}

	r.running = true

	// start goroutine in wg so that graceful shutdown is possible using r.Wait()
	r.wg.Go(func() {

		// run while pending is true, with a debounce delay between runs
		for {
			timer := time.NewTimer(r.debounceDuration)
			select {
			// exit on context cancellation
			case <-r.ctx.Done():
				timer.Stop()

				r.mu.Lock()
				r.running = false
				r.mu.Unlock()

				return
			case <-timer.C:
			}

			// exit if no pending runs
			r.mu.Lock()
			if !r.pending {
				r.running = false
				r.mu.Unlock()
				return
			}

			r.pending = false
			r.mu.Unlock()

			// run the function
			if err := r.runFunc(r.ctx); err != nil && r.ctx.Err() == nil {
				r.logger.Error(r.runFuncStr+" failed", "error", err)
			}
		}
	})
}

// Wait waits for the coalescing runner's goroutine to finish.
func (r *CoalescingRunner) Wait() {
	r.wg.Wait()
}
