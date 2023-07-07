package graceful

import (
	"context"
)

// Logger specifies a logging interface
type Logger interface {
	Info(ctx context.Context, msg string, kv ...interface{})
}
