package lyslog

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func testHandlerOptions(level slog.Level) *slog.HandlerOptions {
	return &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}
}

func TestSplitStreamHandler_RoutesByLevel(t *testing.T) {
	var stdout, stderr bytes.Buffer
	h := NewSplitStreamHandler(&stdout, &stderr, testHandlerOptions(slog.LevelDebug))
	logger := slog.New(h)

	logger.Debug("debug-message")
	logger.Info("info-message")
	logger.Warn("warn-message")
	logger.Error("error-message")

	stdoutLog := stdout.String()
	stderrLog := stderr.String()

	if !strings.Contains(stdoutLog, "debug-message") {
		t.Fatalf("expected debug log in stdout, got: %q", stdoutLog)
	}
	if !strings.Contains(stdoutLog, "info-message") {
		t.Fatalf("expected info log in stdout, got: %q", stdoutLog)
	}
	if !strings.Contains(stdoutLog, "warn-message") {
		t.Fatalf("expected warn log in stdout, got: %q", stdoutLog)
	}
	if strings.Contains(stdoutLog, "error-message") {
		t.Fatalf("did not expect error log in stdout, got: %q", stdoutLog)
	}

	if !strings.Contains(stderrLog, "error-message") {
		t.Fatalf("expected error log in stderr, got: %q", stderrLog)
	}
	if strings.Contains(stderrLog, "info-message") || strings.Contains(stderrLog, "warn-message") || strings.Contains(stderrLog, "debug-message") {
		t.Fatalf("did not expect non-error logs in stderr, got: %q", stderrLog)
	}
}

func TestSplitStreamHandler_EnabledRespectsConfiguredLevel(t *testing.T) {
	h := NewSplitStreamHandler(&bytes.Buffer{}, &bytes.Buffer{}, testHandlerOptions(slog.LevelWarn))

	if h.Enabled(context.Background(), slog.LevelInfo) {
		t.Fatal("expected info level to be disabled")
	}
	if !h.Enabled(context.Background(), slog.LevelWarn) {
		t.Fatal("expected warn level to be enabled")
	}
	if !h.Enabled(context.Background(), slog.LevelError) {
		t.Fatal("expected error level to be enabled")
	}
}

func TestSplitStreamHandler_WithAttrsAppliesToBothStreams(t *testing.T) {
	var stdout, stderr bytes.Buffer
	h := NewSplitStreamHandler(&stdout, &stderr, testHandlerOptions(slog.LevelDebug))
	logger := slog.New(h).With("service", "lys-ref")

	logger.Info("info-with-attr")
	logger.Error("error-with-attr")

	if !strings.Contains(stdout.String(), "service=lys-ref") {
		t.Fatalf("expected attrs in stdout log, got: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "service=lys-ref") {
		t.Fatalf("expected attrs in stderr log, got: %q", stderr.String())
	}
}

func TestSplitStreamHandler_WithGroupAppliesToBothStreams(t *testing.T) {
	var stdout, stderr bytes.Buffer
	h := NewSplitStreamHandler(&stdout, &stderr, testHandlerOptions(slog.LevelDebug))
	logger := slog.New(h).WithGroup("request")

	logger.Info("grouped-info", "id", 101)
	logger.Error("grouped-error", "id", 202)

	if !strings.Contains(stdout.String(), "request.id=101") {
		t.Fatalf("expected grouped key in stdout log, got: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "request.id=202") {
		t.Fatalf("expected grouped key in stderr log, got: %q", stderr.String())
	}
}
