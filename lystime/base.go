package lystime

import (
	"context"
	"time"
)

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
