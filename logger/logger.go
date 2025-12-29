// Package logger defines the logging interface for delish.
package logger

import "context"

// Logger is the interface consumers must implement for delish logging.
type Logger interface {
	Info(ctx context.Context, msg string, kv ...any)
	Trace(ctx context.Context, msg string, kv ...any)
	Error(ctx context.Context, msg string, err error, kv ...any)
	WithFields(ctx context.Context, kv ...any) context.Context
	SetLevel(ctx context.Context, level string) (err error)
	GetLevel() string
}
