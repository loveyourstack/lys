package lysexec

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSuccess(t *testing.T) {

	res, err := Run(context.Background(), "sh", RunOptions{}, "-c", "printf 'hello'")

	require.NoError(t, err)
	assert.Equal(t, "hello", res.Stdout)
	assert.Equal(t, "", res.Stderr)
	assert.Equal(t, 0, res.ExitCode)
	assert.False(t, res.TimedOut)
	assert.False(t, res.Canceled)
	assert.GreaterOrEqual(t, res.DurationMs, int64(0))
}

func TestRunMaxCaptureBytes(t *testing.T) {

	res, err := Run(context.Background(), "sh", RunOptions{MaxCaptureBytes: 5}, "-c", "printf 'hello world'")

	require.NoError(t, err)
	assert.Equal(t, "hello", res.Stdout)
}

func TestRunCanceledByContext(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	res, err := Run(ctx, "sleep", RunOptions{}, "5")

	require.Error(t, err)
	assert.False(t, res.TimedOut)
	assert.True(t, res.Canceled)
	assert.NotEqual(t, 0, res.ExitCode)
}

func TestRunTimedOut(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	res, err := Run(ctx, "sleep", RunOptions{}, "5")

	require.Error(t, err)
	assert.True(t, res.TimedOut)
	assert.False(t, res.Canceled)
	assert.NotEqual(t, 0, res.ExitCode)
}
