package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// ---------------------------------------------------------------------------
// Logger wraps slog.Logger with audit event support and dual-write output.
// ---------------------------------------------------------------------------

// Logger provides structured logging with automatic sensitive-value redaction
// and audit event emission. All output is written simultaneously to both
// os.Stdout and the configured log file.
type Logger struct {
	*slog.Logger
	file *os.File
}

// InitLogger creates a Logger that writes JSON-structured log lines to both
// os.Stdout and the file at logpath. The file is opened in append-only mode.
// Returns an error if the file cannot be opened, which should cause startup
// to fail.
func InitLogger(logpath string) (*Logger, error) {
	f, err := os.OpenFile(logpath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("cannot open log file %q: %w", logpath, err)
	}

	dual := io.MultiWriter(os.Stdout, f)
	base := slog.NewJSONHandler(dual, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	rh := NewRedactingHandler(base)

	return &Logger{
		Logger: slog.New(rh),
		file:   f,
	}, nil
}

// NewTestLogger creates a Logger that writes to the supplied writer only,
// useful for capturing log output in tests without touching the filesystem.
func NewTestLogger(w io.Writer) *Logger {
	base := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	rh := NewRedactingHandler(base)
	return &Logger{
		Logger: slog.New(rh),
	}
}

// Close releases the underlying log file handle. After Close, file writes
// will fail but console output may continue (runtime file failure tolerance).
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// AuditEvent emits a structured audit log entry at INFO level. The event
// name is recorded in the "event" key. Additional key-value pairs should
// follow the audit field convention: method, path, actor, target, op,
// outcome, status, duration_ms.
func (l *Logger) AuditEvent(event string, fields ...any) {
	attrs := make([]any, 0, len(fields)+2)
	attrs = append(attrs, "event", event)
	attrs = append(attrs, fields...)
	l.Info("audit", attrs...)
}

// ---------------------------------------------------------------------------
// Redaction helpers
// ---------------------------------------------------------------------------

// RedactValue returns RedactedPlaceholder if key matches any sensitive key
// (case-insensitive comparison), otherwise returns val unchanged.
func RedactValue(key, val string) string {
	lower := strings.ToLower(key)
	for _, sk := range SensitiveKeys {
		if strings.ToLower(sk) == lower {
			return RedactedPlaceholder
		}
	}
	return val
}

// isSensitiveKey reports whether key matches a known sensitive key name.
func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	for _, sk := range SensitiveKeys {
		if strings.ToLower(sk) == lower {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// RedactingHandler wraps an slog.Handler and redacts sensitive attribute
// values before they reach the underlying handler.
// ---------------------------------------------------------------------------

// RedactingHandler is an slog.Handler that intercepts log records and replaces
// values of sensitive keys with RedactedPlaceholder before delegating to the
// wrapped handler.
type RedactingHandler struct {
	inner slog.Handler
}

// NewRedactingHandler wraps an existing slog.Handler with redaction logic.
func NewRedactingHandler(inner slog.Handler) *RedactingHandler {
	return &RedactingHandler{inner: inner}
}

// Enabled delegates to the inner handler.
func (h *RedactingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

// Handle redacts sensitive attributes in the record before delegating.
func (h *RedactingHandler) Handle(ctx context.Context, r slog.Record) error {
	redacted := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	r.Attrs(func(a slog.Attr) bool {
		redacted.AddAttrs(redactAttr(a))
		return true
	})
	return h.inner.Handle(ctx, redacted)
}

// WithAttrs returns a new handler with the given pre-redacted attributes.
func (h *RedactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	ra := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		ra[i] = redactAttr(a)
	}
	return &RedactingHandler{inner: h.inner.WithAttrs(ra)}
}

// WithGroup returns a new handler with the given group name.
func (h *RedactingHandler) WithGroup(name string) slog.Handler {
	return &RedactingHandler{inner: h.inner.WithGroup(name)}
}

// redactAttr replaces the value of an attribute if its key is sensitive.
// Group attributes are recursively redacted.
func redactAttr(a slog.Attr) slog.Attr {
	if a.Value.Kind() == slog.KindGroup {
		attrs := a.Value.Group()
		redacted := make([]slog.Attr, len(attrs))
		for i, ga := range attrs {
			redacted[i] = redactAttr(ga)
		}
		return slog.Attr{Key: a.Key, Value: slog.GroupValue(redacted...)}
	}

	if isSensitiveKey(a.Key) {
		return slog.String(a.Key, RedactedPlaceholder)
	}
	return a
}
