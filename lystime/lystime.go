package lystime

import (
	"context"
	"errors"
	"time"
)

// RunEvery runs fn every specified duration until the context is canceled.
// fn is not invoked once ctx is canceled.
func RunEvery(ctx context.Context, every time.Duration, fn func()) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if ctx.Err() != nil {
				return
			}
			fn()
		}
	}
}

// Sleep sleeps for the specified duration or until the context is canceled, whichever comes first.
// It is a context-aware replacement for time.Sleep.
func Sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer func() {
		if !t.Stop() {
			select {
			case <-t.C:
			default:
			}
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// WaitOrTimeout blocks until done is closed or d elapses. On timeout it returns an error with msg.
func WaitOrTimeout(done <-chan struct{}, d time.Duration, msg string) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-done:
		return nil
	case <-timer.C:
		return errors.New(msg)
	}
}
