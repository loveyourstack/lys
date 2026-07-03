package lysexec

import (
	"context"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestCoalescingRunner_Trigger_CoalescesToSingleExtraRun(t *testing.T) {
	ctx := t.Context()

	allowFirstRunCh := make(chan struct{})
	firstRunStartedCh := make(chan struct{})
	secondRunStartedCh := make(chan struct{})

	var callCount atomic.Int32
	runner := NewCoalescingRunner(
		ctx,
		func(ctx context.Context) error {
			n := callCount.Add(1)
			switch n {
			case 1:
				close(firstRunStartedCh)
				<-allowFirstRunCh
			case 2:
				close(secondRunStartedCh)
			}
			return nil
		},
		"testRun",
		10*time.Millisecond,
		testLogger(),
	)

	runner.Trigger()

	select {
	case <-firstRunStartedCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for first run to start")
	}

	// While run 1 is in-flight, many triggers should still collapse into one extra run.
	for range 10 {
		runner.Trigger()
	}

	close(allowFirstRunCh)

	select {
	case <-secondRunStartedCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for second run to start")
	}

	time.Sleep(350 * time.Millisecond)

	if got := callCount.Load(); got != 2 {
		t.Fatalf("expected exactly 2 runs (coalesced), got %d", got)
	}

	runner.Wait()
}

func TestCoalescingRunner_Wait_BlocksUntilRunCompletes(t *testing.T) {
	ctx := t.Context()

	releaseRunCh := make(chan struct{})
	runStartedCh := make(chan struct{})

	runner := NewCoalescingRunner(
		ctx,
		func(ctx context.Context) error {
			close(runStartedCh)
			<-releaseRunCh
			return nil
		},
		"testRun",
		10*time.Millisecond,
		testLogger(),
	)

	runner.Trigger()

	select {
	case <-runStartedCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for run to start")
	}

	waitDoneCh := make(chan struct{})
	go func() {
		runner.Wait()
		close(waitDoneCh)
	}()

	select {
	case <-waitDoneCh:
		t.Fatal("wait returned before run completed")
	case <-time.After(200 * time.Millisecond):
	}

	close(releaseRunCh)

	select {
	case <-waitDoneCh:
	case <-time.After(2 * time.Second):
		t.Fatal("wait did not return after run completed")
	}
}

func TestCoalescingRunner_TriggerAfterCancel_DoesNothing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var callCount atomic.Int32
	runner := NewCoalescingRunner(
		ctx,
		func(ctx context.Context) error {
			callCount.Add(1)
			return nil
		},
		"testRun",
		10*time.Millisecond,
		testLogger(),
	)

	runner.Trigger()

	time.Sleep(200 * time.Millisecond)
	runner.Wait()

	if got := callCount.Load(); got != 0 {
		t.Fatalf("expected 0 runs after canceled context, got %d", got)
	}
}

func TestNewCoalescingRunner_ConstructsWithProvidedValues(t *testing.T) {
	ctx := t.Context()
	logger := testLogger()
	runFunc := func(context.Context) error { return nil }

	runner := NewCoalescingRunner(ctx, runFunc, "testRun", 10*time.Millisecond, logger)

	if runner == nil {
		t.Fatal("expected non-nil runner")
	}
	if runner.ctx != ctx {
		t.Fatal("expected context to be assigned")
	}
	if runner.runFunc == nil {
		t.Fatal("expected runFunc to be assigned")
	}
	if runner.runFuncStr != "testRun" {
		t.Fatalf("expected runFuncStr to be testRun, got %q", runner.runFuncStr)
	}
	if runner.debounceDuration != 10*time.Millisecond {
		t.Fatalf("expected debounceDuration to be 10ms, got %s", runner.debounceDuration)
	}
	if runner.logger != logger {
		t.Fatal("expected logger to be assigned")
	}
}
