package lystime

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunEveryRunsUntilCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	calls := make(chan struct{}, 10)
	done := make(chan struct{})

	go func() {
		RunEvery(ctx, 10*time.Millisecond, func() {
			calls <- struct{}{}
		})
		close(done)
	}()

	for range 2 {
		select {
		case <-calls:
		case <-time.After(time.Second):
			t.Fatal("RunEvery did not invoke the function")
		}
	}

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("RunEvery did not stop after cancellation")
	}
}

func TestRunEveryCanceledBeforeFirstTick(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	called := false
	done := make(chan struct{})
	go func() {
		RunEvery(ctx, time.Millisecond, func() {
			called = true
		})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("RunEvery did not stop after cancellation")
	}

	assert.False(t, called)
}

func TestSleepSuccess(t *testing.T) {
	t.Parallel()

	start := time.Now()
	err := Sleep(context.Background(), 20*time.Millisecond)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, elapsed, 15*time.Millisecond)
}

func TestSleepCanceledBeforeStart(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	err := Sleep(ctx, 200*time.Millisecond)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Less(t, elapsed, 100*time.Millisecond)
}

func TestSleepCanceledDuringWait(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(25*time.Millisecond, cancel)

	start := time.Now()
	err := Sleep(ctx, time.Second)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.GreaterOrEqual(t, elapsed, 20*time.Millisecond)
	assert.Less(t, elapsed, 300*time.Millisecond)
}

func TestSleepDeadlineExceeded(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := Sleep(ctx, time.Second)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.GreaterOrEqual(t, elapsed, 20*time.Millisecond)
	assert.Less(t, elapsed, 300*time.Millisecond)
}

func TestWaitOrTimeoutDone(t *testing.T) {
	t.Parallel()

	done := make(chan struct{})
	close(done)

	require.NoError(t, WaitOrTimeout(done, time.Second, "timed out"))
}

func TestWaitOrTimeoutTimeout(t *testing.T) {
	t.Parallel()

	err := WaitOrTimeout(make(chan struct{}), 10*time.Millisecond, "timed out")

	require.Error(t, err)
	assert.EqualError(t, err, "timed out")
}
