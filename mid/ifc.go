package mid

import "context"

// Logger specifies a logging interface
type Logger interface {
	Info(ctx context.Context, msg string, kv ...interface{})
	Error(ctx context.Context, msg string, err error, kv ...interface{})
	WithFields(ctx context.Context, kv ...interface{}) context.Context
}
