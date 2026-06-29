package lyslog

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// SplitStreamHandler is a custom slog.Handler that splits log output between stdout and stderr based on log level.
type SplitStreamHandler struct {
	stdout   slog.Handler
	stderr   slog.Handler
	errLevel slog.Leveler
}

// NewSplitStreamHandler creates a new SplitStreamHandler.
func NewSplitStreamHandler(stdoutWriter, stderrWriter io.Writer, opts *slog.HandlerOptions) *SplitStreamHandler {
	if stdoutWriter == nil {
		stdoutWriter = os.Stdout
	}
	if stderrWriter == nil {
		stderrWriter = os.Stderr
	}
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	return &SplitStreamHandler{
		stdout:   slog.NewTextHandler(stdoutWriter, opts),
		stderr:   slog.NewTextHandler(stderrWriter, opts),
		errLevel: slog.LevelError,
	}
}

func (h *SplitStreamHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if level >= h.errLevel.Level() {
		return h.stderr.Enabled(ctx, level)
	}
	return h.stdout.Enabled(ctx, level)
}

func (h *SplitStreamHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level >= h.errLevel.Level() {
		return h.stderr.Handle(ctx, r)
	}
	return h.stdout.Handle(ctx, r)
}

func (h *SplitStreamHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SplitStreamHandler{
		stdout:   h.stdout.WithAttrs(attrs),
		stderr:   h.stderr.WithAttrs(attrs),
		errLevel: h.errLevel,
	}
}

func (h *SplitStreamHandler) WithGroup(name string) slog.Handler {
	return &SplitStreamHandler{
		stdout:   h.stdout.WithGroup(name),
		stderr:   h.stderr.WithGroup(name),
		errLevel: h.errLevel,
	}
}
