package lysexec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// RunOptions controls command execution and output handling for Run.
type RunOptions struct {

	// command execution options from stdlib
	Dir   string
	Env   []string
	Stdin io.Reader

	// additional options for output handling
	StdoutWriter    io.Writer
	StderrWriter    io.Writer
	TeeStdout       bool // if true, also write stdout to os.Stdout
	TeeStderr       bool // if true, also write stderr to os.Stderr
	MaxCaptureBytes int  // <= 0 means unlimited capture
}

// RunResult contains captured command execution details.
type RunResult struct {
	Stdout     string
	Stderr     string
	ExitCode   int
	DurationMs int64
	TimedOut   bool
	Canceled   bool
}

// Run executes a command with richer options and returns structured output.
func Run(ctx context.Context, name string, opts RunOptions, args ...string) (res RunResult, err error) {

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = opts.Dir
	cmd.Env = opts.Env
	cmd.Stdin = opts.Stdin

	stdoutCapture := newCaptureBuffer(opts.MaxCaptureBytes)
	stderrCapture := newCaptureBuffer(opts.MaxCaptureBytes)

	stdoutWriters := []io.Writer{stdoutCapture}
	if opts.StdoutWriter != nil {
		stdoutWriters = append(stdoutWriters, opts.StdoutWriter)
	}
	if opts.TeeStdout {
		stdoutWriters = append(stdoutWriters, os.Stdout)
	}
	cmd.Stdout = io.MultiWriter(stdoutWriters...)

	stderrWriters := []io.Writer{stderrCapture}
	if opts.StderrWriter != nil {
		stderrWriters = append(stderrWriters, opts.StderrWriter)
	}
	if opts.TeeStderr {
		stderrWriters = append(stderrWriters, os.Stderr)
	}
	cmd.Stderr = io.MultiWriter(stderrWriters...)

	startedAt := time.Now()
	err = cmd.Run()
	res.DurationMs = time.Since(startedAt).Milliseconds()

	res.Stdout = stdoutCapture.String()
	res.Stderr = stderrCapture.String()

	if err == nil {
		res.ExitCode = 0
		return res, nil
	}

	res.ExitCode = -1
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		res.ExitCode = exitErr.ExitCode()
	}

	res.TimedOut = errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded)
	res.Canceled = !res.TimedOut && (errors.Is(ctx.Err(), context.Canceled) || errors.Is(err, context.Canceled))

	stderrMsg := strings.TrimSpace(res.Stderr)
	if stderrMsg != "" {
		return res, fmt.Errorf("command %q failed: %w; stderr: %s", name, err, stderrMsg)
	}

	return res, fmt.Errorf("command %q failed: %w", name, err)
}

type captureBuffer struct {
	max int
	b   bytes.Buffer
}

func newCaptureBuffer(max int) *captureBuffer {
	return &captureBuffer{max: max}
}

func (c *captureBuffer) Write(p []byte) (int, error) {
	if c.max <= 0 {
		return c.b.Write(p)
	}

	remaining := c.max - c.b.Len()
	if remaining <= 0 {
		return len(p), nil
	}

	if len(p) > remaining {
		_, _ = c.b.Write(p[:remaining])
		return len(p), nil
	}

	return c.b.Write(p)
}

func (c *captureBuffer) String() string {
	return c.b.String()
}
