// Package logger configures the application's structured logger on top of
// the standard library's log/slog: JSON output in production, a readable
// text format in development, and context-aware helpers for attaching a
// request-scoped logger to a context.Context.
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

type ctxKey struct{}

// Config controls how the root logger is constructed.
type Config struct {
	// Level is one of "debug", "info", "warn", "error". Defaults to "info".
	Level string
	// Format is one of "json" or "console". Defaults to "console".
	Format string
	// Output defaults to os.Stdout when nil.
	Output io.Writer
}

// New builds a *slog.Logger according to cfg.
func New(cfg Config) *slog.Logger {
	output := cfg.Output
	if output == nil {
		output = os.Stdout
	}

	opts := &slog.HandlerOptions{Level: parseLevel(cfg.Level)}

	var handler slog.Handler
	if strings.EqualFold(cfg.Format, "json") {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithContext returns a new context.Context carrying logger, retrievable via
// FromContext. Use this to attach request-scoped attributes (request ID,
// trace ID) to every log line emitted while handling a request.
func WithContext(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, log)
}

// FromContext returns the logger stored by WithContext, or slog.Default()
// when none is present.
func FromContext(ctx context.Context) *slog.Logger {
	if log, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok && log != nil {
		return log
	}

	return slog.Default()
}
